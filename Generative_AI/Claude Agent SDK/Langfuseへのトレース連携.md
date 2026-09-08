- https://langfuse.com/integrations/frameworks/claude-agent-sdk

- 複数のやり取りを１つのトレースとしてまとめるためには `async with ClaudeSDKClient() as client:`を`with langfuse.start_as_current_observation`で囲む必要がある

  ```python
  import asyncio
  from claude_agent_sdk import query

  import asyncio
  from claude_agent_sdk import (
      ClaudeSDKClient,
      AssistantMessage,
      TextBlock,
      ResultMessage,
      ClaudeAgentOptions,
      HookMatcher

  )
  from dotenv import load_dotenv

  load_dotenv()

  from langfuse import get_client

  langfuse = get_client()

  # Verify connection
  if langfuse.auth_check():
      print("Langfuse client is authenticated and ready!")
  else:
      print("Authentication failed. Please check your credentials and host.")

  from openinference.instrumentation.claude_agent_sdk import ClaudeAgentSDKInstrumentor

  ClaudeAgentSDKInstrumentor().instrument()

  async def main():
      with langfuse.start_as_current_observation(name="conversation", as_type="span"): ⭐️ここ
          async with ClaudeSDKClient() as client:
              # First question
              await client.query("What's the capital of France?")

              # Process response
              async for message in client.receive_response():
                  if isinstance(message, AssistantMessage):
                      for block in message.content:
                          if isinstance(block, TextBlock):
                              print(f"Claude: {block.text}")

              # Follow-up question - the session retains the previous context
              await client.query("What's the population of that city?")

              async for message in client.receive_response():
                  if isinstance(message, AssistantMessage):
                      for block in message.content:
                          if isinstance(block, TextBlock):
                              print(f"Claude: {block.text}")

              # Another follow-up - still in the same conversation
              await client.query("What are some famous landmarks there?")

              async for message in client.receive_response():
                  if isinstance(message, AssistantMessage):
                      for block in message.content:
                          if isinstance(block, TextBlock):
                              print(f"Claude: {block.text}")

  if __name__ == "__main__":
      asyncio.run(main())
  ```