# logseq-export

Tool to export raw [Logseq](https://github.com/logseq/logseq) Markdown files (with `public::` page property) into Markdown blog posts with front matter.

- Takes Logseq page properties (`title:: Hello world`) and turns them into [Front Matter properties](https://gohugo.io/content-management/front-matter/) `title: Hello World`.
- Changes the Markdown syntax to remove the top-level bullet points.

## Install

- Download the latest binary for your OS in the [Releases](https://github.com/viktomas/logseq-export/releases) page
- `go install github.com/viktomas/logseq-export@latest` if you have Go installed

## Usage

### Command

```
logseq-export
  -blogFolder string
        [MANDATORY] Folder where all public pages are exported.
  -graphPath string
        [MANDATORY] Path to the root of your logseq graph containing /pages and /journals directories.
  -assetsRelativePath
        relative path within blogFolder where the assets (images) should be stored (e.g. 'static/images/logseq'). Default is logseq-images (default "logseq-images")
  -webAssetsPathPrefix
    	  path that the images are going to be served on on the web (e.g. '/public/images/logseq'). Default is /logseq-images (default "/logseq-images")
  -unquotedProperties string
        comma-separated list of logseq page properties that won't be quoted in the markdown frontmatter, e.g. 'date,public,slug'
  --listProperties string
        comma-separated list of logseq page properties that will be converted to a list in the mardown frontmatter, e.g. 'tags,series'
```

#### Command example

This is how I run the command on my machine:

```sh
logseq-export \
  --graphPath /Users/tomas/workspace/private/notes \
  --blogFolder /Users/tomas/workspace/private/blog \
  --unquotedProperties date,slug,public,tags \
  --assetsRelativePath static/images/logseq \
  --webAssetsPathPrefix /images/logseq
```

This will take my logseq notes and copies them to blog, it will also copy all the images to `/Users/tomas/workspace/private/blog/static/images/logseq`, but the image links themselves are going to have `/images/logseq` prefix (`![alt](/images/logseq/image.png)`).

### Logseq page properties with a special meaning (all optional)

- `public` - as soon as this page property is present (regardless of value), the page gets exported
- `slug` used as a file name
- `date` it's used as a file name prefix
- `folder` the page is going to be exported in this subfolder e.g. `content/posts`
  - the `folder` property always uses `/` (forward slash) but on Windows, it gets translated to `\` in folder path
  - if the base export folder is `a` and the `folder` page property is `b/c`, then the resulting page will be in `a/b/c` folder
- `image` The value of this property behaves the same way as all Markdown images.
  - if the `image` property contains `../assets/post-image.jpg`, and we run the `logseq-extract` with `--webAssetsPathPrefix /images/logseq -assetsRelativePath static/images/logseq` flags, the resulting Markdown post will have front-matter attribute `image: /images/logseq/post-image.jpg` and the image will be copied to `static/images/logseq/post-image.jpg` in the blog folder.

## From

![logseq test page](./docs/assets/logseq-teset-page.png)

## To

`content/posts/2022-09-25-test-page.md` :

~~~md
---
date: 2022-09-25
categories: "category"
public: true
slug: test-page
folder: "content/posts"
---

This is an example paragraph

- Second level means bullet points

- `logseq-export` also supports multi-level bullet points

```ts
const v = "Hello world"
```

You can
also
have

Multi-line strings
~~~
