# Vite とは

> [!NOTE]
> **一言で**：フロントエンドの **開発サーバー ＆ ビルドツール**。
> 「開発中は爆速で画面を確認でき、本番用には最適化して固める」ための土台。
> React / Vue / Svelte など**フレームワーク非依存**で使える。読み方は「ヴィート」（仏語で"速い"）。

---

## 1. 何をするツールか

ブラウザは TypeScript や JSX、たくさんに分割された `import` をそのまま効率よくは扱えない。
そこで**開発と本番の2つの場面**で面倒を見てくれるのが Vite。

| 場面 | Vite がやること | 使うコマンド |
|---|---|---|
| **開発中** | 開発サーバーを立て、保存した瞬間に画面へ即反映（HMR） | `npm run dev` |
| **本番用** | TS/JSXを変換し、最適化・圧縮して静的ファイルに固める（ビルド） | `npm run build` |

> [!NOTE]
> Vite は**フロントの土台とビルドだけ**を担う。SSRやルーティングは含まない（それらが要るなら Next.js）。
> → [レンダリング方式(SPA・SSR・SSG等).md](./レンダリング方式(SPA・SSR・SSG等).md) / [ディレクトリ構造.md](./ディレクトリ構造.md)

---

## 2. なぜ「速い」のか（最大の特徴）

### 開発中：ネイティブ ESM を使う
従来ツール（webpack 等）は、開発サーバー起動時に **アプリ全体を1つにまとめて（バンドルして）** から表示していた。
規模が大きいほど起動が遅くなる。

Vite は違う：

```
従来 (webpack): 全ファイルをまとめてから起動 → 大規模だと遅い
Vite         : まとめずに、ブラウザが必要としたファイルだけ都度変換して渡す
               → プロジェクト規模に関係なく起動が一瞬
```

- ブラウザ標準の **ES Modules（import/export）** をそのまま活かす。
- 変換が必要な部分だけを**オンデマンド**で処理するので起動が速い。
- 内部で **esbuild**（Go製で非常に高速）を使って変換している。

### 保存 → 即反映（HMR: Hot Module Replacement）
コードを保存すると、**変更した部分だけ**を画面に差し替える。ページ全体をリロードしないので、
入力中のフォーム状態などを保ったまま見た目だけ更新される。

### 本番ビルド：Rollup でまとめる
本番では逆に「バンドルした方が速い」ため、**Rollup** というツールで1つ（数個）にまとめ、
圧縮・最適化（ツリーシェイキング＝未使用コード除去など）した静的ファイルを出力する。

> [!TIP]
> **開発 = バンドルしない（速い）／本番 = バンドルする（最適化）** と、
> 場面ごとに最適な方式を使い分けているのが Vite の肝。

---

## 3. 旧来ツールとの違い

| | Vite | webpack | Create React App (CRA) |
|---|---|---|---|
| 開発起動 | 一瞬（バンドルしない） | 遅くなりがち | 遅い（内部はwebpack） |
| 設定 | シンプル | 複雑になりやすい | 隠蔽されていて弄りにくい |
| 立ち位置 | 現在の主流 | 老舗・今も現役 | **非推奨（開発終了状態）** |

> [!CAUTION]
> かつて React 入門の定番だった **Create React App (CRA) は現在非推奨**。
> 新規で React SPA を作るなら **Vite が事実上の標準**。

---

## 4. 主要コマンド

```bash
# 1. プロジェクト作成（雛形生成）
npm create vite@latest my-app -- --template react-ts

# 2. 依存インストール
cd my-app && npm install

# 3. 開発サーバー起動（http://localhost:5173）
npm run dev

# 4. 本番用にビルド（dist/ に成果物が出る）
npm run build

# 5. ビルド結果をローカルで確認（本番相当の動作チェック）
npm run preview
```

- `--template` には `react-ts`（React+TS）のほか `vue-ts`, `svelte-ts` なども指定できる。
- `dev` と `build` は `package.json` の `scripts` に定義されている（中身は `vite` / `vite build`）。

---

## 5. `vite.config.ts` の役割

Vite の設定ファイル。プラグイン追加やパスエイリアス、開発サーバーのポート/プロキシ等を設定する。

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],           // React を使うためのプラグイン
  resolve: {
    alias: { "@": "/src" },     // import { x } from "@/..." を有効化
  },
  server: {
    port: 3000,                 // 開発サーバーのポート
    proxy: {                    // API 呼び出しをバックエンドに転送（CORS回避）
      "/api": "http://localhost:8080",
    },
  },
});
```

> [!TIP]
> `server.proxy` を使うと、フロント(`localhost:5173`)からの `/api/...` を
> バックエンド(`localhost:8080`)に中継でき、開発時の CORS 問題を避けられる。
> 社内ツールでフロント/バックを別々に立てる構成で便利。

---

## 6. React 以外でも使える

Vite は特定フレームワーク専用ではない。プラグインを差し替えることで各種に対応：

- React … `@vitejs/plugin-react`
- Vue … `@vitejs/plugin-vue`
- Svelte, Solid, Preact, Lit なども対応
- フレームワークなしの素の TS/JS プロジェクトでも使える

---

## まとめ

- Vite は **「開発中は爆速の開発サーバー、本番は最適化ビルド」**を担うフロントの土台。
- 速さの理由：**開発はバンドルせずネイティブESM＋esbuild、本番はRollupでバンドル**。
- **CRA は非推奨、新規 React SPA は Vite が標準**。
- SSR/ルーティングは持たない（要るなら Next.js）。フレームワーク非依存。
