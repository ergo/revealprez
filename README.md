# revealprez

Simple presentation builder using RevealJS and go.

You need to place an index.md file in your presentation directory that has the contents of slides
separated using `----SLIDE----` separator (embedding files is also possible). 

Consult the demo project in `presentation` directory.

The build artifact will contain whole presentation in a form suitable for online serving via any HTTP server.
You can also use builtin server for your presentation

Usage:

First build the binary either directly or by `build.sh` script (min. go version required is 1.16). Sh script will build
binaries for windows, linux and OSX.

    # build the presentation
    ./revealprez_linux_amd64 build --input-dir=presentation
    # will generate `presentation_out` directory for serving
    # You can now serve the presentation
    ./revealprez_linux_amd64 serve

# License

MIT License
