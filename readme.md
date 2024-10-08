# Deadsniper

Dead link (broken link) means a web page you have previously linked to is no longer available.
These dead links have a negative impact on SEO, Security and user experience.
This tool, inspired by [DeadFinder](https://github.com/hahwul/deadfinder), makes it easy to hunt them down.

Note that this tool is much more limited in scope, only allowing use with a sitemap and only checking http/https links.
For customization and a slower, but broader tool, use [DeadFinder](https://github.com/hahwul/deadfinder).

## Usage

```
Usage: deadsniper [options] <link to sitemap.xml>

Options:
  -h | --help    print this help text
  -V | --version print the version number
  -s | --strict  allow only HTTP 200 response codes
  -t | --timeout set the request timeout in seconds (default 5)

Examples:
  deadsniper https://port19.xyz/sitemap.xml
  deadsniper -V
  deadsniper --strict https://port19.xyz/sitemap.xml
  deadsniper -t 1 -s https://port19.xyz/sitemap.xml
```

## Usage via Github Actions

I recommend you download and run the binary, as building a whole docker container slows things down drastically

```
steps:
- name: Run Deadsniper
  run: |
    wget -q "https://github.com/port19x/deadsniper/releases/download/v1.4/deadsniper" && chmod +x ./deadsniper && ./deadsniper https://port19.xyz/sitemap.xml
```

Alternatively, you can use the classical github actions approach at a 10-20s speed cost

```
steps:
- name: Run Deadsniper
  uses: port19x/deadsniper@master
  with:
    sitemap: "https://port19.xyz/sitemap.xml"
```
