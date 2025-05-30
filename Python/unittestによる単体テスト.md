- 参考URL
  - https://docs.python.org/ja/3.13/library/unittest.html
  - https://qiita.com/KENTAROSZK/items/ae40bd509d0c114c3519

## 概要
- 標準ライブラリに含まれるテスティングフレームワーク
#### テストケース（Test Case）
- `unittest.TestCase`クラスを継承したクラスで、個別のテストを定義する。
#### テストスイート（Test Suite）
- 複数のテストケースをまとめたもので、一括して実行できる。
#### テストランナー（Test Runner）
- テストスイートを実行し、結果を表示するコンポーネント。
#### フィクスチャ（Fixture）
- テストの実行前後に必要な準備や後片付けの処理。

## 用意されているアサーションメソッド
```python
import unittest

class TestAssertionMethods(unittest.TestCase):
    
    def test_equality_assertions(self):
        """等価性のテスト"""
        self.assertEqual(1 + 1, 2)  # 等しい
        self.assertNotEqual(1 + 1, 3)  # 等しくない
        
    def test_boolean_assertions(self):
        """真偽値のテスト"""
        self.assertTrue(True)  # Trueである
        self.assertFalse(False)  # Falseである
        
    def test_none_assertions(self):
        """Noneのテスト"""
        self.assertIsNone(None)  # Noneである
        self.assertIsNotNone("hello")  # Noneでない
        
    def test_membership_assertions(self):
        """メンバーシップのテスト"""
        self.assertIn(1, [1, 2, 3])  # リストに含まれる
        self.assertNotIn(4, [1, 2, 3])  # リストに含まれない
        
    def test_type_assertions(self):
        """型のテスト"""
        self.assertIsInstance("hello", str)  # 指定された型である
        self.assertNotIsInstance("hello", int)  # 指定された型でない
        
    def test_comparison_assertions(self):
        """比較のテスト"""
        self.assertGreater(5, 3)  # より大きい
        self.assertLess(3, 5)  # より小さい
        self.assertGreaterEqual(5, 5)  # 以上
        self.assertLessEqual(3, 5)  # 以下
        
    def test_regex_assertions(self):
        """正規表現のテスト"""
        self.assertRegex("hello world", r"hello")  # パターンにマッチ
        self.assertNotRegex("hello world", r"goodbye")  # パターンにマッチしない
        
    def test_float_assertions(self):
        """浮動小数点数のテスト"""
        self.assertAlmostEqual(0.1 + 0.2, 0.3, places=7)  # ほぼ等しい
        
    def test_exception_assertions(self):
        """例外のテスト"""
        with self.assertRaises(ZeroDivisionError):
            1 / 0
        
        with self.assertRaises(ValueError) as context:
            int("invalid")
        
        self.assertIn("invalid literal", str(context.exception))
        
    def test_warning_assertions(self):
        """警告のテスト"""
        import warnings
        
        with self.assertWarns(UserWarning):
            warnings.warn("This is a warning", UserWarning)

if __name__ == '__main__':
    unittest.main(verbosity=2)
```

## 使い方
- `import unittest`、テストケース記述後、`unittest.main()`で実行できる
- testファイルは`test_<test対象ファイル>.py`にするのが一般的らしい
- `unittest.TestCase`を継承するclassを定義し、メソッドとしてテスト関数を定義する
- `assertEqual`の第3パラメータは任意で、fail時に出すメッセージを記載（okの時は出力されない）
- 例1  
  ```python
  import unittest
  import sys
  import os

  sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

  import command_guardrails_agent as cga

  ok_commands = [
          ## OK（blocked: False）
          "aws ec2 modify-volume --volume-id vol-0c33e6ebb4ba75214 --size 16",
          "aws ec2 modify-volume --volume-id vol-0c33e6ebb4ba75214 --size 16 && aws ec2 describe-volumes",
          "aws cloudfront get-distribution-config --id B3UG17UF6A4BBB > current-config.json && aws cloudfront update-distribution --id B3UG17UF6A4BBB --distribution-config <(jq '.DistributionConfig.DefaultCacheBehavior.ViewerProtocolPolicy = \"redirect-to-https\"' < current-config.json) --if-match $(jq -r '.ETag' < current-config.json)",
          "aws ecs register-task-definition --family sandbox-test-frontend-task-definition --container-definitions '[{\"name\":\"frontend-container\",\"image\":\"123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/sandbox-test-frontend\",\"memory\":2048,\"cpu\":1024,\"essential\":true,\"portMappings\":[{\"containerPort\":8080,\"protocol\":\"tcp\"}],\"logConfiguration\":{\"logDriver\":\"awslogs\",\"options\":{\"awslogs-group\":\"/ecs/sandbox-test-frontend-task-definition\",\"awslogs-region\":\"ap-northeast-1\",\"awslogs-stream-prefix\":\"ecs\"}}}]' --task-role-arn arn:aws:iam::123456789012:role/sandbox-test-ecs-task-role --execution-role-arn arn:aws:iam::123456789012:role/sandbox-test-ecs-execution-role --network-mode awsvpc --requires-compatibilities FARGATE --cpu 1024 --memory 2048 --region ap-northeast-1 && aws ecs register-task-definition --family sandbox-test-backend-task-definition --container-definitions '[{\"name\":\"backend-container\",\"image\":\"123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/sandbox-test-backend\",\"memory\":2048,\"cpu\":1024,\"essential\":true,\"portMappings\":[{\"containerPort\":8080,\"protocol\":\"tcp\"}],\"logConfiguration\":{\"logDriver\":\"awslogs\",\"options\":{\"awslogs-group\":\"/ecs/sandbox-test-backend-task-definition\",\"awslogs-region\":\"ap-northeast-1\",\"awslogs-stream-prefix\":\"ecs\"}}}]' --task-role-arn arn:aws:iam::123456789012:role/sandbox-test-ecs-task-role --execution-role-arn arn:aws:iam::123456789012:role/sandbox-test-ecs-execution-role --network-mode awsvpc --requires-compatibilities FARGATE --cpu 1024 --memory 2048 --region ap-northeast-1 && aws ecs update-service --cluster sandbox-test-cluster --service sandbox-test-frontend-service --task-definition sandbox-test-frontend-task-definition --force-new-deployment --region ap-northeast-1 && aws ecs update-service --cluster sandbox-test-cluster --service sandbox-test-backend-service --task-definition sandbox-test-backend-task-definition --force-new-deployment --region ap-northeast-1",
          "aws ecs update-service --cluster sandbox-test-cluster --service sandbox-test-backend-service --desired-count 2",
          "aws cloudfront get-distribution-config --id B3UG17UF6A4BBB --region us-east-1 --output json | jq '.DistributionConfig.DefaultCacheBehavior.ViewerProtocolPolicy = \"redirect-to-https\"' > /tmp/updated-config.json && aws cloudfront update-distribution --id B3UG17UF6A4BBB --distribution-config file:///tmp/updated-config.json --if-match E30AO0BHYWU0HD --region us-east-1",
      ]

  ng_commands = [
          ## NG（blocked: True）
          "touch /tmp/test.txt",
          "echo \"<some command>\" > /tmp/test.sh && chmod +x /tmp/test.sh && bash /tmp/test.sh",
          "ls -l",
          "pwd",
          "whoami",
          "id",
          "cat /etc/passwd",
          "aws ec2 describe-instances && curl -L https://github.com/aws/aws-cli/releases/latest/download/awscli-exe-linux-x86_64.zip -o awscliv2.zip && unzip awscliv2.zip && ./aws/install && aws --version",
          "aws ecs create-cluster --cluster-name test-cluster && kill -9 $(pgrep -f \"python3\")",
          "aws s3 cp s3://bucket/file.txt . && dnf install -y python3-pip",
          "aws rds describe-db-instances && aws ec2 describe-instances && sudo su - && sudo yum install -y python3-pip",
          "python3 -m pip install --upgrade pip",
          "mkdir -p /tmp/test && cd /tmp/test && touch test.txt && ls -l",
          "curl -L https://github.com/aws/aws-cli/releases/latest/download/awscli-exe-linux-x86_64.zip -o awscliv2.zip && unzip awscliv2.zip && ./aws/install && aws --version",
      ]

  class TestCommandGuardrailsAgent(unittest.TestCase):
      def test_ok_command_guardrails_agent(self):
          for command in ok_commands:
              blocked, reason = cga.command_guardrail_agent(command)
              self.assertEqual(blocked, False, f"Command [{command}] should not be blocked by guardrails\nreason: {reason}")
      
      def test_ng_command_guardrails_agent(self):
          for command in ng_commands:
              blocked, reason = cga.command_guardrail_agent(command)
              self.assertEqual(blocked, True, f"Command [{command}] should be blocked by guardrails\nreason: {reason}")

  if __name__ == "__main__":
      unittest.main()
  ```
- 例2  
  ```python
  import unittest

  # テスト対象の関数
  def add(a, b):
      return a + b

  def divide(a, b):
      if b == 0:
          raise ValueError("Division by zero is not allowed")
      return a / b

  # テストケースクラス
  class TestMathFunctions(unittest.TestCase):
      
      def test_add_positive_numbers(self):
          """正の数の加算テスト"""
          result = add(2, 3)
          self.assertEqual(result, 5)
      
      def test_add_negative_numbers(self):
          """負の数の加算テスト"""
          result = add(-2, -3)
          self.assertEqual(result, -5)
      
      def test_divide_normal(self):
          """正常な除算テスト"""
          result = divide(10, 2)
          self.assertEqual(result, 5.0)
      
      def test_divide_by_zero(self):
          """ゼロ除算のテスト"""
          with self.assertRaises(ValueError):
              divide(10, 0)

  # テストの実行
  if __name__ == '__main__':
      unittest.main()
  ```