# FastAPIのファイルアップロード（UploadFile）について

FastAPIでファイルを受け取るには `UploadFile`（と`File()`）、通常のフォームフィールドを受け取るには `Form()` を使う。どちらも `multipart/form-data` 形式のリクエストを前提にしており、これはJSONボディ（→[PydanticのBaseModelを使用してRequest Body内のJSONパラメータを受け取る方法](PydanticのBaseModelを使用してRequest%20Body内のJSONパラメータを受け取る方法.md)）とは**別のエンコーディング**である点がポイント。

> 出典: [FastAPI公式 - Request Files](https://fastapi.tiangolo.com/tutorial/request-files/) / [Request Forms](https://fastapi.tiangolo.com/tutorial/request-forms/)

---

## 1. 前提: `python-multipart` が必要

```bash
pip install python-multipart
```

`UploadFile` / `File()` / `Form()` は内部で `multipart/form-data` をパースするために `python-multipart` パッケージに依存する。入れ忘れると起動時または実行時にエラーになる。

---

## 2. `UploadFile` の基本

```python
from fastapi import FastAPI, UploadFile

app = FastAPI()


@app.post("/upload")
async def upload_file(file: UploadFile):
    contents = await file.read()          # bytesとして読み込む（非同期）
    return {
        "filename": file.filename,
        "content_type": file.content_type,
        "size": len(contents),
    }
```

| 属性/メソッド | 内容 |
|---|---|
| `file.filename` | クライアントが送ってきた元のファイル名 |
| `file.content_type` | MIMEタイプ（例: `image/png`） |
| `file.size` | ファイルサイズ（バイト、Starlette 0.24+） |
| `await file.read(size=-1)` | 中身を非同期で読む（サイズ指定でチャンク読みも可） |
| `await file.write(data)` | 書き込み（あまり使わない） |
| `await file.seek(offset)` | 読み取り位置を移動 |
| `await file.close()` | クローズ |
| `file.file` | 内部の`SpooledTemporaryFile`（同期I/Oライブラリに渡したい場合） |

### `bytes` との違い

`File()` を素の `bytes` 型と組み合わせても受け取れるが、`UploadFile` の方が実務向き。

| 項目 | `bytes` | `UploadFile` |
|---|---|---|
| 保持場所 | **全量メモリに読み込む** | 一定サイズまでメモリ、それ以降はディスクのスプールファイル |
| 大きいファイル | メモリを圧迫しやすい | 効率的（`SpooledTemporaryFile`使用） |
| メタデータ（ファイル名等） | 取れない | `filename` / `content_type` が取れる |
| 非同期読み込み | 不要（すでにメモリ上） | `await file.read()` が必要 |

```python
from fastapi import File

@app.post("/upload-bytes")
async def upload_bytes(file: bytes = File()):
    return {"size": len(file)}
```

→ 小さいファイル・単純な用途以外は基本的に `UploadFile` を使う。

---

## 3. 複数ファイルのアップロード

```python
@app.post("/upload-multiple")
async def upload_multiple(files: list[UploadFile]):
    return {"filenames": [f.filename for f in files]}
```

---

## 4. ファイルとフォームフィールドを同時に受け取る

```python
from typing import Annotated
from fastapi import Form, UploadFile

@app.post("/upload-with-meta")
async def upload_with_meta(
    file: UploadFile,
    description: Annotated[str, Form()],
    is_public: Annotated[bool, Form()] = False,
):
    return {"filename": file.filename, "description": description, "is_public": is_public}
```

> [!WARNING]
> **ファイルアップロードがあるエンドポイントでは、Pydanticモデル（JSONボディ）と`File`/`Form`を混在できない**。`multipart/form-data` の中に他のフィールドを含めたいときは、上記のように個々のフィールドを `Form()` で受け取る必要がある（1つのPydanticモデルとしてまとめて受けることはできない）。

---

## 5. 受け取ったファイルを保存する

```python
import shutil
from pathlib import Path

UPLOAD_DIR = Path("uploads")
UPLOAD_DIR.mkdir(exist_ok=True)


@app.post("/upload-save")
async def upload_save(file: UploadFile):
    dest = UPLOAD_DIR / file.filename
    with dest.open("wb") as buffer:
        shutil.copyfileobj(file.file, buffer)   # 同期I/O（file.file は同期ファイルオブジェクト）
    return {"saved_to": str(dest)}
```

> [!NOTE]
> `shutil.copyfileobj` は同期処理。`async def`のエンドポイント内でこれを直接使うと、大きいファイルではイベントループを塞ぐ可能性がある（→[async def と def の使い分けについて](async%20def%20と%20def%20の使い分けについて.md)）。重いファイルI/Oが多い場合は、エンドポイントを`def`にする、または`await run_in_threadpool(...)`で逃がすことを検討する。

```python
from starlette.concurrency import run_in_threadpool

await run_in_threadpool(shutil.copyfileobj, file.file, dest.open("wb"))
```

---

## 6. バリデーション（サイズ・拡張子・MIMEタイプ）

FastAPI自体はファイルサイズや拡張子を自動検証しないため、必要なら自分でチェックする。

```python
from fastapi import HTTPException

ALLOWED_TYPES = {"image/png", "image/jpeg"}
MAX_SIZE = 5 * 1024 * 1024  # 5MB


@app.post("/upload-image")
async def upload_image(file: UploadFile):
    if file.content_type not in ALLOWED_TYPES:
        raise HTTPException(status_code=415, detail="Unsupported file type")

    contents = await file.read()
    if len(contents) > MAX_SIZE:
        raise HTTPException(status_code=413, detail="File too large")

    return {"filename": file.filename, "size": len(contents)}
```

> [!WARNING]
> `file.content_type` はクライアントが申告した値であり、**偽装可能**。厳密にファイル種別を検証したい場合は、実際のバイト列（マジックナンバー）を見るライブラリ（`python-magic`等）を使う。

---

## 7. ストリーミングでアップロードを受ける（巨大ファイル向け）

`await file.read()` はデフォルトで全量読み込む。巨大ファイルはチャンク単位で処理する。

```python
CHUNK_SIZE = 1024 * 1024  # 1MB


@app.post("/upload-stream")
async def upload_stream(file: UploadFile):
    dest = UPLOAD_DIR / file.filename
    with dest.open("wb") as f:
        while chunk := await file.read(CHUNK_SIZE):
            f.write(chunk)
    return {"saved_to": str(dest)}
```

- レスポンス側の分割送信は [StreamingResponseについて](%20StreamingResponseについて.md) を参照（アップロード＝受信側、StreamingResponse＝送信側という対になる関係）

---

## ポイントまとめ

- ファイルは `UploadFile`（`File()`と組み合わせも可）、通常フィールドは `Form()` で受け取る。どちらも `multipart/form-data` が前提で `python-multipart` が必要
- `UploadFile` はメモリ節約（スプールファイル）とメタデータ（`filename`/`content_type`）取得の面で `bytes` より実務向き
- ファイルアップロードのあるエンドポイントでは、JSONボディ（Pydanticモデル）と同時受信はできない。フィールドは個別に`Form()`で宣言する
- 保存処理の同期I/O（`shutil.copyfileobj`等）は大きいファイルだとイベントループを塞ぎうる。`def`化か`run_in_threadpool`を検討
- サイズ・MIMEタイプの検証はFastAPIが自動でやらないため自前でチェックする。`content_type`は偽装可能な点に注意
- 巨大ファイルは`await file.read(CHUNK_SIZE)`でチャンク読みする
