package assets

import (
	"path/filepath"

	"github.com/inconshreveable/log15"

	"polydawn.net/repeatr/def"
	"polydawn.net/repeatr/io"
	"polydawn.net/repeatr/io/transmat/cachedir"
	"polydawn.net/repeatr/io/transmat/tar"
)

var assets = map[string]integrity.CommitID{
	"runc": integrity.CommitID("asdf"),
}

/*
	Gets a path to the rootfs of the named asset.  The asset
	may be fetched if it's not available.

	Usage might be like

		cmd.Path = filepath.Join(assets.Get("runc"), "bin/runc")

	There is no versioning information in parameters because
	this is where the buck stops: a build of repeatr was tested against
	and shall use exactly one known version of a thing.  The assets
	dirs will themselves be treated like CAS, of course: multiple
	installs of different versions of repeatr on the system may share
	an assets cache without fuss.
*/
func Get(assetName string) string {

	// Note: haven't got an API that proxies all the monitoring options yet.
	// Be nice to have that someday, but tbh we need to develop the core of that further first.

	arena := transmat().Materialize(
		integrity.TransmatKind("tar"),
		assets[assetName],
		[]integrity.SiloURI{
			"http+ca://repeatr.s3.amazonaws.com/assets/",
		},
		log15.New(log15.DiscardHandler), // this is foolish, but i just feel Wrong requiring a logger as an arg to `asset.Get`.
	)

	return arena.Path()
}

/*
	A separate transmat is used for the asset system.

	Assets use a separate cache.

	There's also only one kind of transmat enabled here -- it really
	only makes sense for the asset system to use the tar transmat,
	since that's so easily bundled without large extraneous components.
*/
func transmat() integrity.Transmat {
	workDir := filepath.Join(def.Base(), "assets")
	dirCacher := cachedir.New(filepath.Join(workDir, "cache"), map[integrity.TransmatKind]integrity.TransmatFactory{
		integrity.TransmatKind("tar"): tar.New,
	})
	return dirCacher
}
