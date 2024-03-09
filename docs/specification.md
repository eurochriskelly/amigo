# Watcher

The watcher program will simply watch for file changes in a selection of directories and server those changes over http.

## API

For example, when a user connects to the service, e.g.

http://localhost:9191/registry.json

They will receive a list of available files in the following JSON structure:

    [
      {
        "label": "directory1/foo",
        "type": "xqy",
        "url": "http://localhost:9191/files/directory1/foo.xqy"
      },
      etc.
    ]

The when the user follows one of the URLs provided, they will get the latest contents of that file.

## Launching

To launch the tool the user should run as follows:

    ./watcher --directory foo:/path/to/foo --directory bar:/path/to/bar --directory /path/to/qux  --extensions xqy,js,sjs

Each directory is provided as a path or an label:path pair. The label is used
when provided. Otherwise the last path part of the directory is used. For
example, if the local file system had the following contents.

    foo
    └── bar
            ├── baz.xqy
            └── qux.xqy

and if the tool was started as follows:
   
    ./watcher --directory /foo/bar --extensions xqy

The tool should generate the following out put for the registry.json endpoint:

    [
      {
        "label": "bar/baz",
        "type": "xqy",
        "url": "files/foo/bar/baz.xqy"
      },
      {
        "label": "bar/qux",
        "type": "xqy",
        "url": "files/foo/bar/qux.xqy"
      }
    ]
		    
