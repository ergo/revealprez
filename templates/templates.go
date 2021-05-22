package templates

const IndexTemplate = `<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">

    <title>Presentation</title>

    <link rel="stylesheet" href="dist/reveal.css">
    <link rel="stylesheet" href="dist/theme/black.css">

    <!-- Theme used for syntax highlighting of code -->
    <link rel="stylesheet" href="dist/theme/zenburn.css">

</head>
<body>
<div class="reveal">
    <div class="slides">
        {{- range $index, $element := . -}} {{ $element.RenderedSlide }}
		{{- end -}}
    </div>
</div>

<script src="dist/reveal.js"></script>
<script src="plugin/markdown/markdown.js"></script>
<script src="plugin/highlight/highlight.js"></script>

<script>
    // More info about config & dependencies:
    // - https://github.com/hakimel/reveal.js#configuration
    // - https://github.com/hakimel/reveal.js#dependencies
    Reveal.initialize({
        slideNumber: true,
        showNotes: true,
        plugins: [ RevealMarkdown, RevealHighlight ]
    });
</script>
</body>
</html>`
