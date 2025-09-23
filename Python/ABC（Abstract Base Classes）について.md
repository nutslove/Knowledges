## ABC（Abstract Base Classes）とは
- Pythonの標準ライブラリ`abc`モジュールで提供される、抽象的な（共通の）インターフェースを定義するための仕組み
- サブクラスで必ず実装しなければならないメソッドを指定できる（サブクラスに「必ず実装しなければならないメソッド」を強制できる）。抽象メソッドが実装されていないサブクラスはインスタンス化できない。
- 抽象基底クラスで`ABC`を継承し、`@abstractmethod`デコレータを使用して抽象メソッドを定義する
  - サブクラスは抽象基底クラスを継承し、`@abstractmethod`で定義されたすべての抽象メソッドを実装する必要がある
- 例  
  ```python
  from abc import ABC, abstractmethod
  from typing import List

  ## 抽象基底クラス
  class DataProcessor(ABC):
      @abstractmethod
      def load_data(self, source: str) -> List:
          pass
      
      @abstractmethod
      def process_data(self, data: List) -> List:
          pass
      
      @abstractmethod
      def save_data(self, data: List, destination: str) -> None:
          pass
      
      # テンプレートメソッド（通常のメソッドも定義可能）
      def execute(self, source: str, destination: str):
          data = self.load_data(source)
          processed_data = self.process_data(data)
          self.save_data(processed_data, destination)

  ## サブクラス
  class CSVProcessor(DataProcessor):
      def load_data(self, source: str) -> List:
          # CSV読み込み処理
          return ["csv", "data"]
      
      def process_data(self, data: List) -> List:
          # CSV用の処理
          return [item.upper() for item in data]
      
      def save_data(self, data: List, destination: str) -> None:
          # CSV保存処理
          print(f"Saving {data} to {destination}")
  ```