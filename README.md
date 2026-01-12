# vb-code-go

## 1. 汎用圧縮（ファイル圧縮）

| 技術 | 特徴 | 用途 |
| --- | --- | --- |
| gzip | 広く普及、バランス型 | Webコンテンツ、ログ |
| zstd | 高速＋高圧縮率、現在の主流 | Facebook開発、多用途 |
| lz4 | 超高速、圧縮率は控えめ | リアルタイム処理 |
| brotli | 高圧縮率 | Web配信（Google開発） |

## 2. 整数列専用圧縮（VB Codeの仲間）

| 技術 | 特徴 | 主な採用 |
| --- | --- | --- |
| VB Code | シンプル、実装容易 | 教育、軽量システム |
| PForDelta | ブロック単位で高速 | Lucene（過去） |
| SIMD-BP128 | CPU命令で超高速 | 現代の検索エンジン |
| Roaring Bitmap | 集合演算に最適化 | Elasticsearch、Spark |

## 3. 列指向DB専用

| 技術 | 特徴 |
| --- | --- |
| Run-Length Encoding | 連続する同じ値を圧縮 |
| Dictionary Encoding | 文字列を整数IDに変換 |
| Delta Encoding | 差分保存（VB Codeと併用多い） |

---

## 使い分けの指針

| やりたいこと | 推奨技術 |
| --- | --- |
| ファイルを小さくしたい | zstd（現在の王道） |
| Web配信を高速化 | brotli / gzip |
| ログを圧縮 | zstd / lz4 |
| 検索インデックス | Roaring Bitmap |
| 整数列を圧縮（学習用） | VB Code |
| リアルタイム圧縮 | lz4 / snappy |
