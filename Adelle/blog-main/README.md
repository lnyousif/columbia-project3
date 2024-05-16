# The Confident Scientist

![Purple head](./public/images/purple_head.webp)

> [The definitive compendium of cutting edge computer science](https://housker.github.io/confidentscientist/)

## Table of Contents

1. [Overview](#overview)
1. [Usage](#usage)
1. [Script](#script)
1. [Requirements](#requirements)
1. [System](#system)
1. [Design](#design)
1. [Resources](#resources)
<!-- 1. [Test](#test)
1. [Features](#features)
1. [Issues](#issues)
1. [Contributing](#contributing)
1. [License](#license) -->

## Overview
Static blog and portfolio that builds in a millisecond and loads just as fast.

## Usage

From the root directory, add posts with `hugo new posts/"$(date +%y-%m-%d).md"`

Compress images `cwebp -q 50 <old name>.jpeg -o <new name>.webp`, and add them to _static/images_

Build them with `hugo -t ananke`

`cd public` to see the generated files. Commit and push them. The site will automatically updates with a few seconds' delay.

To preview changes locally, run `hugo serve`.

To add additional [themes](https://themes.gohugo.io/), `git submodule add <url of theme repository> themes/<theme name>` and update _hugo.toml_ according to the directions on the theme's readme.

## Usage

I've added a couple scripts to ~/.zshrc to streamline the creation and publication of blog posts.

To create a new blog post:
```shell
function blog {
    local blog_dir="/Users/adellehousker/fun/website/cs/blog/"
    if [[ "$PWD/" == "$blog_dir" ]]; then
        local file_name="$(date +%y-%m-%d).md"
        if [ ! -e "$file_name" ]; then
            hugo new content posts/$file_name
            touch "$blog_dir/scratch/$file_name"
        fi
        code "$blog_dir/content/en/posts/$file_name"
    else
        echo "Command only available at $blog_dir"
    fi
}
```

To make the post live:
```shell
function publish {
    local blog_dir="/Users/adellehousker/fun/website/cs/blog"
    if [[ "$PWD" == "$blog_dir" ]]; then
        local fnames=""
        for file in $blog_dir/content/en/posts/*; do
        local fname=`basename $file`
        local fnames="${fnames} $fname";
        ./translate/translate $fname;
        done
        echo $fnames
        hugo -t ananke
        cd public
        git add .
        git commit -m "posting${fnames}"
        git push origin main
    else
        echo "Command only available at $blog_dir"
    fi
}
```

## Requirements

Go v1.14+

[Hugo 0.125.5](https://formulae.brew.sh/formula/hugo)

[webp 1.4.0](https://formulae.brew.sh/formula/webp)

## System

You'll need two repos: One that holds the Hugo code and content. Create that with `hugo new site blog`. `cd blog` and set up dependency management `hugo mod init "github.com/housker/blog.git"` and version control `git init`. Add a theme `git submodule add https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke`

The second repo will hold the generated assets to be deployed to Github Pages. When you run `hugo -t ananke` it will generate files into the _public_ directory. Make _public_ a git submodule by running `git submodule add -b main https://github.com/housker/confidentscientist.git public` so those files can be pushed and directly deployed from there. You can verify by `cd`-ing into _public_ and running `git remote -v`.

## Design

Static site generators (SSG) create websites that are optimized for search engines, relatively secure from vulnerabilities, and fast to load. They are appropriate for use cases where no user-specific customization or real-time updates are needed. Next.js, Gatsby, and Hugo are examples of SSGs. Hugo was chosen primarily because build time is [significantly faster](https://draft.dev/learn/hugo-vs-gatsby)

## Resources

- [Hugo](https://gohugo.io/)
- [Ryan Schachte's tutorial](https://www.youtube.com/watch?v=LIFvgrRxdt4)
- [WebP image compression](https://web.dev/articles/codelab-serve-images-webp)
- Photo by <a href="https://unsplash.com/@dynamicwang?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">Dynamic Wang</a> on <a href="https://unsplash.com/photos/a-womans-head-is-shown-with-a-purple-background-CKfqRX9l52g?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">Unsplash</a>
