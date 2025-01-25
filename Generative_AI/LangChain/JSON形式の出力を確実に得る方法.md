- Fuction Callingは実際に関数を呼び出さずに、JSON形式のデータを確実に出力させる用途でも良く使われる

## `with_structured_output`
- https://python.langchain.com/docs/how_to/structured_output/
- `with_structured_output`を使ってLLMに確実にJSON形式で出力させることができる
  - `with_structured_output`に対応しているモデル
    - https://python.langchain.com/docs/integrations/chat/
- 設定例  
  ```python
  from typing import Optional
  from pydantic import BaseModel, Field

  # Pydantic
  class Joke(BaseModel):
      """Joke to tell user."""

      setup: str = Field(description="The setup of the joke")
      punchline: str = Field(description="The punchline to the joke")
      rating: Optional[int] = Field(
          default=None, description="How funny the joke is, from 1 to 10"
      )


  structured_llm = llm.with_structured_output(Joke)

  structured_llm.invoke("Tell me a joke about cats")
  ```