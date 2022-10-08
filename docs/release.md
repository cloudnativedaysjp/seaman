# release コマンド

## Summary

`release` コマンドを呼び出すと、リポジトリとリリースレベルをそれぞれ選択することになります。これらを選択すると `release/major` などのラベルの付与された PR を自動生成するのがこのコマンドの責務です。

注意点として、当コマンドの責務は上述した PR の作成のみです。以下は別途 GitHub Actions を利用して実現する必要があります。

* Bot により作成された PR を merge した際に自動でタグを付与する
* タグが付与された際に production 環境にデプロイする

Setup 手順にこれらの GitHub Actions の用意の手順も記載されているため、ご参照ください。

## Setup

* リリース対象のリポジトリに以下の名前のラベルを作成してください。
    * `release/major`
    * `release/minor`
    * `release/patch`


* seaman のコンフィグにリリース対象のリポジトリ名を追加してください。
    * Bot のコンフィグは dreamkast-infra リポジトリに配置されいています ([link](https://github.com/cloudnativedaysjp/dreamkast-infra/blob/main/manifests/app/seaman/configmap.yaml))

```diff
  # For example
  release:
    targets:
+     - url: https://github.com/ShotaKitazawa/kube-portal
+       baseBranch: master
```

### Bot により作成された PR を merge した際に自動でタグを付与する

* Bot により作成された PR を merge したときに自動でタグをインクリメントする GitHub Action を作成します。 (eg. [push-tag-by-releasebot.yml](https://github.com/cloudnativedaysjp/seaman/blob/main/.github/workflows/push-tag-by-releasebot.yml))
    * `if: contains(github.event.pull_request.title, '[dreamkast-releasebot]')` : releasebot が作成した PR にのみ反応するようにしています
    * `Generate token` step : GitHub Actions から tag が push されたことを契機に別の action をトリガするために、GitHub App のクレデンシャルを利用するようにしています
        * GitHub App は [`GitOps for CloudNativeDays`](https://github.com/organizations/cloudnativedaysjp/settings/installations/29106044) を利用してください
        * `APP_ID` , `PRIVATE_KEY` はそれぞれ GitHub の Actions secrets にて値を登録してください


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
          token: ${{ steps.generate_token.outputs.token }}

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

### タグが付与された際に production 環境にデプロイする

「タグが付与されたらデプロイを実施する」ための Action を用意してください。この Action の内容はアプリケーションのデプロイ方法によって異なります。

* 例1. [cloudnativedaysjp/seaman](https://github.com/cloudnativedaysjp/seaman/blob/main/.github/workflows/gitops-prd.yml)
    * dreamkast-infra リポジトリのマニフェスト更新
* 例2. [cloudnativedaysjp/dreamkast-function](https://github.com/cloudnativedaysjp/dreamkast-functions/blob/main/.github/workflows/deploy-prd.yml)
    * AWS CDK を用いてデプロイ
* 例3. [cloudnativedaysjp/website](https://github.com/cloudnativedaysjp/website/tree/main/.github/workflows)
    * [`AWS Amplify` GitHub App](https://github.com/apps/aws-amplify-ap-northeast-1) により自動でデプロイされるため、デプロイ用の Action は存在しない
