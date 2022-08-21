# release サブコマンド

## 概要

`release` サブコマンドを呼び出すと、リポジトリとリリースレベルをそれぞれ選択することになります。これらを選択すると `release/major` などのラベルの付与された PR を自動生成するのがこのコマンドの責務です。

注意点として、上記により作成された PR を merge しても自動でタグは付与されません。タグを付与するための GitHub Actions を別途用意しなければいけないです。
Setup 手順に GitHub Actions の用意の手順も記載されているためご参照ください。

## Setup

リリース対象のリポジトリを追加する方法についてです。

* 追加したいリポジトリに以下の名前のラベルを作成します。
    * `release/major`
    * `release/minor`
    * `release/patch`

* 上記ラベルが付与された PR を merge したときに自動でタグをインクリメントする GitHub Action を作成します。以下はその例です。

```yaml
name: Push a new tag with Pull Request

on:
  pull_request:
    types: [closed]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - uses: actions-ecosystem/action-release-label@v1
        id: release-label
        if: ${{ github.event.pull_request.merged == true }}

      - uses: actions-ecosystem/action-get-latest-tag@v1
        id: get-latest-tag
        if: ${{ steps.release-label.outputs.level != null }}

      - uses: actions-ecosystem/action-bump-semver@v1
        id: bump-semver
        if: ${{ steps.release-label.outputs.level != null }}
        with:
          current_version: ${{ steps.get-latest-tag.outputs.tag }}
          level: ${{ steps.release-label.outputs.level }}

      - uses: actions-ecosystem/action-push-tag@v1
        if: ${{ steps.release-label.outputs.level != null }}
        with:
          tag: ${{ steps.bump-semver.outputs.new_version }}
          message: '${{ steps.bump-semver.outputs.new_version }}: PR #${{ github.event.pull_request.number }} ${{ github.event.pull_request.title }}'
```

* 上記実施後、bot のコンフィグにリポジトリを追加してください。

```diff
  release:
    targets:
+     - url: https://github.com/ShotaKitazawa/kube-portal
+       baseBranch: master
```
