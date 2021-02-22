# go-appengine-mail-receiver

AppEngine(Go)でメールを受信するサンプル

1. `anything@[GCP_PROJECT].appspotmail.com` あてのメールをAppEngineで受信して `eml` をCloud Storage `RAW_EML_BUCKET` に保存
2. Storageの `onFinalize` でCloud Functions for Firebase `onFinalizeRawEml` を起動、別Bucketにコピー
3. コピー先Bucketの `onFinalize` で `onFinalizeEml` を起動、 `eml` をパースして `json` にして保存
