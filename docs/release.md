# release サブコマンド

## 概要

`release` サブコマンドを呼び出すと、リポジトリとリリースレベルをそれぞれ選択することになります。これらを選択すると `release/major` などのラベルの付与された PR を自動生成するのがこのコマンドの責務です。

注意点として、上記により作成された PR を merge しても自動でタグは付与されません。タグを付与するための GitHub Actions を別途用意しなければいけないです。
Setup 手順に GitHub Actions の用意の手順も記載されているためご参照ください。

## Setup

リリース対象のリポジトリを追加する方法についてです。

* 追加したいリポジトリに以下の名前のラベルを作成してください。
    * `release/major`
    * `release/minor`
    * `release/patch`

* 上記ラベルが付与された PR を merge したときに自動でタグをインクリメントする GitHub Action を作成してください。 (eg. [tagging.yml](https://github.com/cloudnativedaysjp/seaman/blob/main/.github/workflows/tagging.yml))
    * `if: contains(github.event.pull_request.title, '[dreamkast-releasebot]')` : releasebot が作成した PR にのみ反応するようにしています
    * `Generate token` step : GitHub Actions から tag が push されたことを契機に別の action をトリガするために、GitHub App のクレデンシャルを利用するようにしています
        * GitHub App は [`GitOps for CloudNativeDays`](https://github.com/organizations/cloudnativedaysjp/settings/installations/29106044) を利用してください


```yaml
name: Push a new tag with merged Pull Request

on:
  pull_request:
    types: [closed]

jobs:
  tagging:
    runs-on: ubuntu-latest
    if: contains(github.event.pull_request.title, '[dreamkast-releasebot]')
    steps:
      - name: Generate token
        id: generate_token
        uses: tibdex/github-app-token@v1
        with:
          app_id: ${{ secrets.APP_ID }}
          private_key: ${{ secrets.PRIVATE_KEY }}

      - uses: actions/checkout@v3
        with:
          persist-credentials: false

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

      - name: set credential
        env:
          GITHUB_TOKEN: ${{ steps.generate_token.outputs.token }}
        run: |
          git config remote.origin.url "https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}"

      - uses: actions-ecosystem/action-push-tag@v1
        if: ${{ steps.release-label.outputs.level != null }}
        with:
          tag: ${{ steps.bump-semver.outputs.new_version }}
          message: '${{ steps.bump-semver.outputs.new_version }}: PR #${{ github.event.pull_request.number }} ${{ github.event.pull_request.title }}'
```

* 上記実施後、bot のコンフィグにリポジトリを追加してください。
    * Bot のコンフィグは dreamkast-infra リポジトリに配置されいています ([link](https://github.com/cloudnativedaysjp/dreamkast-infra/blob/main/manifests/app/seaman/configmap.yaml))

```diff
  release:
    targets:
+     - url: https://github.com/ShotaKitazawa/kube-portal
+       baseBranch: master
```
