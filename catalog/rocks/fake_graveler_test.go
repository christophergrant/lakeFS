package rocks

import (
	"bytes"
	"context"
	"strings"

	"github.com/treeverse/lakefs/graveler"
)

type FakeGraveler struct {
	KeyValue           map[string]*graveler.Value
	Err                error
	ListIterator       graveler.ValueIterator
	DiffIterator       graveler.DiffIterator
	RepositoryIterator graveler.RepositoryIterator
	BranchIterator     graveler.BranchIterator
}

func fakeGravelerBuildKey(repositoryID graveler.RepositoryID, ref graveler.Ref, key graveler.Key) string {
	return strings.Join([]string{repositoryID.String(), ref.String(), key.String()}, "/")
}

func (g *FakeGraveler) Get(_ context.Context, repositoryID graveler.RepositoryID, ref graveler.Ref, key graveler.Key) (*graveler.Value, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	k := fakeGravelerBuildKey(repositoryID, ref, key)
	v := g.KeyValue[k]
	if v == nil {
		return nil, graveler.ErrNotFound
	}
	return v, nil
}

func (g *FakeGraveler) Set(_ context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, key graveler.Key, value graveler.Value) error {
	if g.Err != nil {
		return g.Err
	}
	k := fakeGravelerBuildKey(repositoryID, graveler.Ref(branchID.String()), key)
	g.KeyValue[k] = &value
	return nil
}

func (g *FakeGraveler) Delete(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, key graveler.Key) error {
	panic("implement me")
}

func (g *FakeGraveler) List(_ context.Context, _ graveler.RepositoryID, _ graveler.Ref) (graveler.ValueIterator, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	return g.ListIterator, nil
}

func (g *FakeGraveler) GetRepository(ctx context.Context, repositoryID graveler.RepositoryID) (*graveler.Repository, error) {
	panic("implement me")
}

func (g *FakeGraveler) CreateRepository(ctx context.Context, repositoryID graveler.RepositoryID, storageNamespace graveler.StorageNamespace, branchID graveler.BranchID) (*graveler.Repository, error) {
	panic("implement me")
}

func (g *FakeGraveler) ListRepositories(ctx context.Context) (graveler.RepositoryIterator, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	return g.RepositoryIterator, nil
}

func (g *FakeGraveler) DeleteRepository(ctx context.Context, repositoryID graveler.RepositoryID) error {
	panic("implement me")
}

func (g *FakeGraveler) CreateBranch(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, ref graveler.Ref) (*graveler.Branch, error) {
	panic("implement me")
}

func (g *FakeGraveler) UpdateBranch(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, ref graveler.Ref) (*graveler.Branch, error) {
	panic("implement me")
}

func (g *FakeGraveler) GetBranch(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID) (*graveler.Branch, error) {
	panic("implement me")
}

func (g *FakeGraveler) GetTag(ctx context.Context, repositoryID graveler.RepositoryID, tagID graveler.TagID) (*graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) CreateTag(ctx context.Context, repositoryID graveler.RepositoryID, tagID graveler.TagID, commitID graveler.CommitID) error {
	panic("implement me")
}

func (g *FakeGraveler) DeleteTag(ctx context.Context, repositoryID graveler.RepositoryID, tagID graveler.TagID) error {
	panic("implement me")
}

func (g *FakeGraveler) ListTags(ctx context.Context, repositoryID graveler.RepositoryID) (graveler.TagIterator, error) {
	panic("implement me")
}

func (g *FakeGraveler) Log(ctx context.Context, repositoryID graveler.RepositoryID, commitID graveler.CommitID) (graveler.CommitIterator, error) {
	panic("implement me")
}

func (g *FakeGraveler) ListBranches(_ context.Context, _ graveler.RepositoryID) (graveler.BranchIterator, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	return g.BranchIterator, nil
}

func (g *FakeGraveler) DeleteBranch(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID) error {
	panic("implement me")
}

func (g *FakeGraveler) Commit(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, committer string, message string, metadata graveler.Metadata) (graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) GetCommit(ctx context.Context, repositoryID graveler.RepositoryID, commitID graveler.CommitID) (*graveler.Commit, error) {
	panic("implement me")
}

func (g *FakeGraveler) Dereference(ctx context.Context, repositoryID graveler.RepositoryID, ref graveler.Ref) (graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) Reset(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID) error {
	panic("implement me")
}

func (g *FakeGraveler) ResetKey(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, key graveler.Key) error {
	panic("implement me")
}

func (g *FakeGraveler) ResetPrefix(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, key graveler.Key) error {
	panic("implement me")
}

func (g *FakeGraveler) Revert(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID, ref graveler.Ref) (graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) Merge(ctx context.Context, repositoryID graveler.RepositoryID, from graveler.Ref, to graveler.BranchID, committer string, message string, metadata graveler.Metadata) (graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) DiffUncommitted(ctx context.Context, repositoryID graveler.RepositoryID, branchID graveler.BranchID) (graveler.DiffIterator, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	return g.DiffIterator, nil
}

func (g *FakeGraveler) Diff(_ context.Context, _ graveler.RepositoryID, _, _ graveler.Ref) (graveler.DiffIterator, error) {
	if g.Err != nil {
		return nil, g.Err
	}
	return g.DiffIterator, nil
}

func (g *FakeGraveler) CommitExistingMetaRange(_ context.Context, _ graveler.RepositoryID, _ graveler.BranchID, _ graveler.MetaRangeID, _ string, _ string, _ graveler.Metadata) (graveler.CommitID, error) {
	panic("implement me")
}

func (g *FakeGraveler) WriteMetaRange(_ context.Context, _ graveler.RepositoryID, _ graveler.ValueIterator) (*graveler.MetaRangeID, error) {
	panic("implement me")
}

type FakeValueIterator struct {
	Data  []*graveler.ValueRecord
	Index int
	Error error
}

func NewFakeValueIterator(data []*graveler.ValueRecord) *FakeValueIterator {
	return &FakeValueIterator{Data: data, Index: -1}
}

func (m *FakeValueIterator) Next() bool {
	if m.Index >= len(m.Data) {
		return false
	}
	m.Index++
	return m.Index < len(m.Data)
}

func (m *FakeValueIterator) SeekGE(id graveler.Key) {
	m.Index = len(m.Data)
	for i, d := range m.Data {
		if bytes.Compare(d.Key, id) >= 0 {
			m.Index = i - 1
			return
		}
	}
}

func (m *FakeValueIterator) Value() *graveler.ValueRecord {
	return m.Data[m.Index]
}

func (m *FakeValueIterator) Err() error {
	return nil
}

func (m *FakeValueIterator) Close() {}

type FakeDiffIterator struct {
	Data  []*graveler.Diff
	Index int
}

func NewFakeDiffIterator(data []*graveler.Diff) *FakeDiffIterator {
	return &FakeDiffIterator{Data: data, Index: -1}
}

func (m *FakeDiffIterator) Next() bool {
	if m.Index >= len(m.Data) {
		return false
	}
	m.Index++
	return m.Index < len(m.Data)
}

func (m *FakeDiffIterator) SeekGE(_ graveler.Key) {
	panic("implement me")
}

func (m *FakeDiffIterator) Value() *graveler.Diff {
	return m.Data[m.Index]
}

func (m *FakeDiffIterator) Err() error {
	return nil
}

func (m *FakeDiffIterator) Close() {}

type FakeRepositoryIterator struct {
	Data  []*graveler.RepositoryRecord
	Index int
}

func NewFakeRepositoryIterator(data []*graveler.RepositoryRecord) *FakeRepositoryIterator {
	return &FakeRepositoryIterator{Data: data, Index: -1}
}

func (m *FakeRepositoryIterator) Next() bool {
	if m.Index >= len(m.Data) {
		return false
	}
	m.Index++
	return m.Index < len(m.Data)
}

func (m *FakeRepositoryIterator) SeekGE(id graveler.RepositoryID) {
	m.Index = len(m.Data)
	for i, repo := range m.Data {
		if repo.RepositoryID >= id {
			m.Index = i - 1
			return
		}
	}
}

func (m *FakeRepositoryIterator) Value() *graveler.RepositoryRecord {
	return m.Data[m.Index]
}

func (m *FakeRepositoryIterator) Err() error {
	return nil
}

func (m *FakeRepositoryIterator) Close() {}

type FakeBranchIterator struct {
	Data  []*graveler.BranchRecord
	Index int
}

func NewFakeBranchIterator(data []*graveler.BranchRecord) *FakeBranchIterator {
	return &FakeBranchIterator{Data: data, Index: -1}
}

func (m *FakeBranchIterator) Next() bool {
	if m.Index >= len(m.Data) {
		return false
	}
	m.Index++
	return m.Index < len(m.Data)
}

func (m *FakeBranchIterator) SeekGE(id graveler.BranchID) {
	m.Index = len(m.Data)
	for i, repo := range m.Data {
		if repo.BranchID >= id {
			m.Index = i - 1
			return
		}
	}
}

func (m *FakeBranchIterator) Value() *graveler.BranchRecord {
	return m.Data[m.Index]
}

func (m *FakeBranchIterator) Err() error {
	return nil
}

func (m *FakeBranchIterator) Close() {}
