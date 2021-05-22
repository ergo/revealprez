# revealprez

Simple presentation builder using RevealJS and go.

You need to place an index.md file in your presentation directory that has the contents of slides
separated using `----SLIDE----` separator (embedding files is also possible). 

Consult the demo project in `presentation` directory.

The build artifact will contain whole presentation in a form suitable for online serving via any HTTP server.
You can also use builtin server for your presentation

Usage:

    # build the presentation
    ./revealprez build --input-dir=presentation
    # will generate `presentation_out` directory for serving
    # You can now serve the presentation
    ./revealprez serve

# License

MIT License
