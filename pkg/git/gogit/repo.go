// Code generated by interfacer; DO NOT EDIT

package gogit

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Repository is an interface generated for "github.com/go-git/go-git/v5.Repository".
type Repository interface {
	BlobObject(plumbing.Hash) (*object.Blob, error)
	BlobObjects() (*object.BlobIter, error)
	Branch(string) (*config.Branch, error)
	Branches() (storer.ReferenceIter, error)
	CommitObject(plumbing.Hash) (*object.Commit, error)
	CommitObjects() (object.CommitIter, error)
	Config() (*config.Config, error)
	ConfigScoped(config.Scope) (*config.Config, error)
	CreateBranch(*config.Branch) error
	CreateRemote(*config.RemoteConfig) (*git.Remote, error)
	CreateRemoteAnonymous(*config.RemoteConfig) (*git.Remote, error)
	CreateTag(string, plumbing.Hash, *git.CreateTagOptions) (*plumbing.Reference, error)
	DeleteBranch(string) error
	DeleteObject(plumbing.Hash) error
	DeleteRemote(string) error
	DeleteTag(string) error
	Fetch(*git.FetchOptions) error
	FetchContext(context.Context, *git.FetchOptions) error
	Head() (*plumbing.Reference, error)
	Log(*git.LogOptions) (object.CommitIter, error)
	Notes() (storer.ReferenceIter, error)
	Object(plumbing.ObjectType, plumbing.Hash) (object.Object, error)
	Objects() (*object.ObjectIter, error)
	Prune(git.PruneOptions) error
	Push(*git.PushOptions) error
	PushContext(context.Context, *git.PushOptions) error
	Reference(plumbing.ReferenceName, bool) (*plumbing.Reference, error)
	References() (storer.ReferenceIter, error)
	Remote(string) (*git.Remote, error)
	Remotes() ([]*git.Remote, error)
	RepackObjects(*git.RepackConfig) error
	ResolveRevision(plumbing.Revision) (*plumbing.Hash, error)
	SetConfig(*config.Config) error
	Tag(string) (*plumbing.Reference, error)
	TagObject(plumbing.Hash) (*object.Tag, error)
	TagObjects() (*object.TagIter, error)
	Tags() (storer.ReferenceIter, error)
	TreeObject(plumbing.Hash) (*object.Tree, error)
	TreeObjects() (*object.TreeIter, error)
	Worktree() (*git.Worktree, error)
}
