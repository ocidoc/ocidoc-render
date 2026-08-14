# OCIDoc renderer benchmark

This fixture exercises ordinary Markdown, GFM, links, media and raw HTML.

## Status

| Check | Result |
| --- | --- |
| Sanitizer | ~~disabled~~ enabled |
| Local asset | ![Diagram](assets/diagram.svg) |

* [Local documentation](guide.md)
* <https://example.com/docs>
* [Contact](mailto:docs@example.com)

<audio controls src="assets/demo.ogg" preload="metadata"></audio>
<video controls poster="https://example.com/poster.png">
  <source src="https://example.com/demo.webm" type="video/webm">
  <source src="assets/demo.mp4" type="video/mp4">
  <track kind="subtitles" src="assets/demo.vtt" srclang="en" label="English">
</video>

<div class="notice" onclick="alert(1)" style="color:red">
  Safe text with an <a href="javascript:alert(1)">unsafe link</a>.
</div>

<script>alert(1)</script>
