# React とは

> [!NOTE]
> **一言で**：**UI（画面）を「コンポーネント」という部品の組み合わせで作るための JavaScript ライブラリ**。
> Meta（旧 Facebook）が開発。Webフロントエンドで最も広く使われている。

---

## 1. 何をするものか

素の JavaScript で画面を作ると、「データが変わったら、どのDOM要素をどう書き換えるか」を
**手作業で全部書く**必要があり、規模が大きくなると破綻する。

React は考え方を逆にする：

> [!IMPORTANT]
> **「今のデータならこういう画面になる」だけを宣言すれば、
> 実際のDOM更新はReactがやってくれる。**

- **命令的（imperative）**：素のJS。「この要素を探して、textを書き換えて…」と手順を書く
- **宣言的（declarative）**：React。「データ = state」と「見た目 = UI」の対応だけを書く

```jsx
// 「count が今いくつなら、こう表示する」を宣言するだけ
function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count} 回クリック</button>;
}
```
→ `count` が変われば、Reactが自動でその部分だけ画面を更新する。

---

## 2. 中心概念

### コンポーネント (Component)
画面を構成する**再利用可能な部品**。関数として書く（＝関数コンポーネント）。
部品を組み合わせて画面を作る（ボタン → フォーム → ページ、のように積み上げる）。

```jsx
function Welcome() {
  return <h1>こんにちは</h1>;
}
```

### JSX
JavaScript の中に HTML のような記法を混ぜられる構文。React の見た目はこれで書く。

```jsx
const element = <h1 className="title">Hello, {userName}</h1>;
```
- `{ }` の中には JavaScript の式を書ける。
- 実体は JS に変換される（HTMLそのものではない）。`class` ではなく `className` など細かな違いあり。

### Props（プロパティ）
親コンポーネントから子へ**データを渡す**仕組み。**上から下への一方向**。

```jsx
function Greeting({ name }) {      // 受け取る側
  return <p>ようこそ、{name}さん</p>;
}
<Greeting name="太郎" />           // 渡す側
```

### State（状態）
コンポーネントが持つ**変化するデータ**。`useState` で扱う。
**stateが変わると、その部分だけ再描画される**のがReactの核。

```jsx
const [count, setCount] = useState(0);
// count      : 現在の値
// setCount   : 値を更新する関数（これを呼ぶと再描画が走る）
```

### Hooks（フック）
関数コンポーネントに機能を追加する仕組み。`use〜` で始まる関数。

| Hook | 役割 |
|---|---|
| `useState` | 状態を持つ |
| `useEffect` | 副作用（API取得・購読・タイマー等）を実行 |
| `useContext` | 離れたコンポーネント間でデータ共有 |
| `useMemo` / `useCallback` | 計算結果・関数をキャッシュして最適化 |
| `useRef` | 再描画を起こさずに値やDOM参照を保持 |

---

## 3. なぜ速い / 何が嬉しいか

### 仮想DOM (Virtual DOM)
Reactは実際のDOMをいきなり触らず、**メモリ上の軽量なコピー（仮想DOM）**でまず差分を計算し、
**変わった箇所だけ**を実DOMに反映する。これで無駄な描画を減らす。

### コンポーネントの再利用
同じ部品（Button, Card…）を各所で使い回せる。修正も1箇所で済む。
→ [ディレクトリ構造.md](./ディレクトリ構造.md) の「Colocation / feature-based」設計につながる。

---

## 4. React 単体では“UIライブラリ”に過ぎない

React が担うのは**画面描画だけ**。実アプリに必要な周辺は別ライブラリを組み合わせる。

| 目的 | 代表的なもの |
|---|---|
| ページ遷移（ルーティング） | React Router / TanStack Router |
| サーバーデータ取得・キャッシュ | TanStack Query (React Query) / SWR |
| グローバル状態管理 | Zustand / Redux / Jotai |
| フレームワーク（SSR等込み） | **Next.js** / Remix |
| ビルド環境 | **Vite**（→ [Viteとは.md](./Viteとは.md)） |

> [!TIP]
> **React vs Next.js**：Next.js は「Reactを土台にSSRやルーティング等を全部入れたフレームワーク」。
> Reactは部品、Next.jsは家一式、というイメージ。
> レンダリング方式の違いは → [レンダリング方式(SPA・SSR・SSG等).md](./レンダリング方式(SPA・SSR・SSG等).md)

---

## 5. 似た立ち位置のもの（比較）

| | 一言 |
|---|---|
| **React** | 部品ベースのUIライブラリ。エコシステム最大 |
| **Vue** | Reactに似た思想。学習しやすいとされる |
| **Svelte** | ビルド時にコンパイル。仮想DOMを使わず軽量 |
| **Angular** | 全部入りの重厚なフレームワーク（Google製） |

---

## 6. 最小の動く例

```jsx
import { useState } from "react";

function App() {
  const [count, setCount] = useState(0);
  return (
    <div>
      <h1>カウンター</h1>
      <p>現在: {count}</p>
      <button onClick={() => setCount(count + 1)}>+1</button>
    </div>
  );
}

export default App;
```

- `useState(0)` で状態を用意
- ボタンを押す → `setCount` → `count` が変わる → 該当部分だけ再描画

---

## まとめ

- Reactは **「データ(state)を宣言すれば、UIを自動で同期してくれる」** UIライブラリ。
- 基本要素：**コンポーネント / JSX / Props / State / Hooks**。
- 仮想DOMで必要な箇所だけ効率的に更新する。
- React単体はUIだけ。ルーティング・データ取得・SSRなどは Next.js や各種ライブラリで補う。
