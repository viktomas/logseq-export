# logseq-export

Tool to export raw [Logseq](https://github.com/logseq/logseq) Markdown pages (with `public::` page property) into Markdown blog posts with front matter.

- Takes Logseq page properties (`title:: Hello world`) and turns them into [Front Matter properties](https://gohugo.io/content-management/front-matter/) `title: Hello World`.
- Changes the Markdown syntax to remove the top-level bullet points.

## Install

- Download the latest binary for your OS in the [Releases](https://github.com/viktomas/logseq-export/releases) page
- `go install github.com/viktomas/logseq-export@latest` if you have Go installed

## Usage

The `logseq-export` utility will export the pages into an export folder that can then be imported into your static site generator.

```mermaid
graph LR;
LS[Logseq graph] --"logseq-export"--> EF[export folder]
EF --"import_to_hugo.sh"--> HU[Hugo static site generator]
```

### Export

```
logseq-export
  -outputFolder string
        [MANDATORY] Folder where all public pages are exported.
  -logseqFolder string
        [MANDATORY] Path to the root of your logseq graph containing /pages and /journals directories.
```

This command also expects you have a file called `export.yaml` in your logseq folder.

TODO: remove the assets paths

```yml
unquotedProperties:
  - date
  - tags
assetsRelativePath: "static/images/logseq"
webAssetsPathPrefix: "/images/logseq"
```

- `assetsRelativePath` relative path within blogFolder where the assets (images) should be stored (e.g. 'static/images/logseq'). Default is logseq-images (default "logseq-images")
- `webAssetsPathPrefix` path that the images are going to be served on on the web (e.g. '/public/images/logseq'). Default is /logseq-images (default "/logseq-images")
- `unquotedProperties` list of logseq page properties that won't be quoted in the markdown frontmatter

#### Command example

This is how I run the command on my machine:

```sh
logseq-export \
  --logseqFolder /Users/tomas/workspace/private/notes \
  --outputFolder /tmp/logseq-export \
```

This will take my logseq notes and copies them to the export folder, it will also copy all the images to `/tmp/logseq-export/logseq-assets`, but the image links themselves are going to have `/logseq-asstes/` prefix (`![alt](/logseq/assets/image.png)`).

#### Constraints

- `logseq-export` assumes that all the pages you want to export are in `pages/` folder inside your `logseqFolder`.


### Import

```sh
# these environment variables are optional
# the values in this example are default values
export BLOG_CONTENT_FODLER="/graph"
export BLOG_IMAGES_FOLDER="/assets/graph"

# copies pages from `/tmp/logseq/export/logseq-pages` to `~/workspace/private/blog/content/graph`
# copies assets from `/tmp/logseq/export/logseq-assets` to `~/workspace/private/blog/static/assets/graph`
# replaces all `/logseq-assets` in all image URLs with `/assets/graph`
./import_to_hugo.sh \
  /tmp/logseq-export
  ~/workspace/private/blog
```

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

`content/graph/2022-09-25-test-page.md` :

~~~md
---
date: 2022-09-25
categories: "category"
public: true
slug: test-page
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
