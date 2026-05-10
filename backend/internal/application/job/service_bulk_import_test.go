package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	domainIdol "github.com/kuro48/idol-api/internal/domain/idol"
	domainJob "github.com/kuro48/idol-api/internal/domain/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryJobRepo struct {
	jobs map[string]*domainJob.Job
}

func newInMemoryJobRepo() *inMemoryJobRepo {
	return &inMemoryJobRepo{jobs: make(map[string]*domainJob.Job)}
}

func (r *inMemoryJobRepo) Save(_ context.Context, j *domainJob.Job) error {
	r.jobs[j.ID()] = j
	return nil
}

func (r *inMemoryJobRepo) FindByID(_ context.Context, id string) (*domainJob.Job, error) {
	job, ok := r.jobs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return job, nil
}

func (r *inMemoryJobRepo) Update(_ context.Context, j *domainJob.Job) error {
	r.jobs[j.ID()] = j
	return nil
}

func (r *inMemoryJobRepo) FindByStatus(_ context.Context, status domainJob.JobStatus, limit int) ([]*domainJob.Job, error) {
	var jobs []*domainJob.Job
	for _, job := range r.jobs {
		if job.Status() == status {
			jobs = append(jobs, job)
		}
		if len(jobs) == limit {
			break
		}
	}
	return jobs, nil
}

type fakeBulkImportIdolPort struct {
	inputs []appIdol.CreateInput
	errFor map[string]error
}

func (f *fakeBulkImportIdolPort) CreateIdol(_ context.Context, item appIdol.CreateInput) (*domainIdol.Idol, error) {
	f.inputs = append(f.inputs, item)
	if err := f.errFor[item.Name]; err != nil {
		return nil, err
	}
	return nil, nil
}

func TestExecuteBulkImport_ImportsAllItems(t *testing.T) {
	t.Parallel()

	repo := newInMemoryJobRepo()
	job := domainJob.NewJob(domainJob.JobTypeBulkImport, []byte(`{"items":[{"name":"星野みく","birthdate":"2001-05-01","agency_id":"agency-1","aliases":["みく"],"tag_ids":["tag-1"]},{"name":"佐藤あい"}]}`), "admin")
	job.SetID("job-import-all")
	repo.jobs[job.ID()] = job

	importer := &fakeBulkImportIdolPort{}
	svc := NewApplicationService(repo, importer)

	svc.executeBulkImport(job.ID(), job.Payload())

	require.Len(t, importer.inputs, 2)
	assert.Equal(t, "星野みく", importer.inputs[0].Name)
	assert.Equal(t, []string{"tag-1"}, importer.inputs[0].TagIDs)
	assert.Equal(t, "佐藤あい", importer.inputs[1].Name)

	storedJob, err := repo.FindByID(context.Background(), job.ID())
	require.NoError(t, err)
	assert.Equal(t, domainJob.JobStatusCompleted, storedJob.Status())

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(storedJob.Result(), &result))
	assert.Equal(t, float64(2), result["processed"])
	assert.Equal(t, float64(2), result["success"])
	assert.Empty(t, result["errors"])
}

func TestExecuteBulkImport_RecordsPartialFailures(t *testing.T) {
	t.Parallel()

	repo := newInMemoryJobRepo()
	job := domainJob.NewJob(domainJob.JobTypeBulkImport, []byte(`{"items":[{"name":"成功"},{"name":"失敗"}]}`), "admin")
	job.SetID("job-import-partial")
	repo.jobs[job.ID()] = job

	importer := &fakeBulkImportIdolPort{
		errFor: map[string]error{
			"失敗": errors.New("duplicate idol"),
		},
	}
	svc := NewApplicationService(repo, importer)

	svc.executeBulkImport(job.ID(), job.Payload())

	require.Len(t, importer.inputs, 2)

	storedJob, err := repo.FindByID(context.Background(), job.ID())
	require.NoError(t, err)
	assert.Equal(t, domainJob.JobStatusCompleted, storedJob.Status())

	var result struct {
		Processed int `json:"processed"`
		Success   int `json:"success"`
		Errors    []struct {
			Index int    `json:"index"`
			Name  string `json:"name"`
			Error string `json:"error"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(storedJob.Result(), &result))
	assert.Equal(t, 2, result.Processed)
	assert.Equal(t, 1, result.Success)
	require.Len(t, result.Errors, 1)
	assert.Equal(t, 1, result.Errors[0].Index)
	assert.Equal(t, "失敗", result.Errors[0].Name)
	assert.Contains(t, result.Errors[0].Error, "duplicate idol")
}
