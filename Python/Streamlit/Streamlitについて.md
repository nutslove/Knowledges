## install
- `pip install streamlit`

## Streamlit実行
- `streamlit run <.pyファイル>`

## MultiPage App
- https://docs.streamlit.io/get-started/tutorials/create-a-multipage-app
- `pages`ディレクトリ配下のpythonファイルから、initial画面用のpythonファイルがある(1つ上の階層)ディレクトリにあるPythonコードから`import`する時、  
  `from .. import`ではなく、同じ階層のPythonファイルをimportする時と同じように`import`だけでimportする