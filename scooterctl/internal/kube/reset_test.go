package kube

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestResetJobResumesOrReplacesFailedAttempt(t *testing.T) {
	for _, failed := range []bool{
		false,
		true,
	} {
		name := "data-reset-test"
		condition := batchv1.JobComplete
		if failed {
			condition = batchv1.JobFailed
		}
		old := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: standNamespace,
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{{
					Type:   condition,
					Status: corev1.ConditionTrue,
				}},
			},
		}
		client := fake.NewSimpleClientset(old)
		created := 0
		client.PrependReactor("create", "jobs", func(action ktesting.Action) (bool, runtime.Object, error) {
			created++
			job := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
			job.Status.Conditions = []batchv1.JobCondition{{
				Type:   batchv1.JobComplete,
				Status: corev1.ConditionTrue,
			}}
			return false, nil, nil
		})
		root, err := filepath.Abs("../../..")
		if err != nil {
			t.Fatal(err)
		}
		cm, job, err := ResetResources(root, name)
		if err != nil {
			t.Fatal(err)
		}
		cluster := &Client{Typed: client}
		if err := cluster.ResetData(t.Context(), cm, job, io.Discard); err != nil {
			t.Fatal(err)
		}
		if (created == 1) != failed {
			t.Fatalf("failed=%t, создано новых Job=%d", failed, created)
		}
		if !strings.Contains(cm.Data["profile.sql"], "insert into public.users") || !strings.Contains(cm.Data["scooter.sql"], "insert into public.scooters") {
			t.Fatal("сброс должен загрузить канонические seed после очистки")
		}
		if err := cluster.CleanupReset(t.Context(), name); err != nil {
			t.Fatal(err)
		}
	}
}

func TestResetReplaysSeedUpWithoutRollback(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "services", "profile", "tools", "migrations")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	body := "-- +goose Up\ninsert into users values (1);\n-- +goose Down\ndelete from users;\n"
	if err := os.WriteFile(filepath.Join(dir, "002_seed_users.sql"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_create_users.sql"), []byte("create table users (id int);"), 0600); err != nil {
		t.Fatal(err)
	}
	sql, err := seedMigrationSQL(root, "profile")
	if err != nil || !strings.Contains(sql, "insert into users") || strings.Contains(sql, "delete") || strings.Contains(sql, "create table") {
		t.Fatalf("unsafe reset SQL: %q, err=%v", sql, err)
	}
	sql, err = seedMigrationSQL(root, "pricing")
	if err != nil || sql != "" {
		t.Fatalf("unexpected SQL for another owner: %q, err=%v", sql, err)
	}
}
