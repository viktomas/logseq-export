##!/bin/bash

set -e

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
    echo "Usage: move_logseq_content.sh <export_folder> <blog_folder>"
    exit 1
fi

# Extract arguments
export_folder="$1"
blog_folder="$2"

# Check if the export folder exists
if [ ! -d "$export_folder" ]; then
    echo "Error: The export folder does not exist."
    exit 1
fi

# Check if the blog folder exists
if [ ! -d "$blog_folder" ]; then
    echo "Error: The blog folder does not exist."
    exit 1
fi

blog_content_folder="${BLOG_CONTENT_FODLER:-/graph}"
images_folder="${BLOG_IMAGES_FOLDER:-/assets/graph}"

# by default the files get copied as follows
# - /logseq-pages -> /content/graph
# - /logseq-assets -> /static/assets/graph
pages_destination="$blog_folder/content$blog_content_folder"
assets_destination="$blog_folder/static$images_folder"

# delete existing pages and assets
rm -rf "$pages_destination"
rm -rf "$assets_destination"

# prepare the directories
mkdir -p "$pages_destination"
mkdir -p "$assets_destination"

# Move the content of logseq-pages to the new destination
cp -R "$export_folder/logseq-pages"/* "$pages_destination/"

# Move the content of logseq-assets to the new destination
cp -R "$export_folder/logseq-assets"/* "$assets_destination/"

# replace the /logseq-asstes/ paths with the hugo image folder
find "$pages_destination" -type f -exec sed -i -e "s@/logseq-assets/@$images_folder/@g" {} \;

echo "Content moved successfully."
