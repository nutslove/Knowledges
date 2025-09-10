- `ToolNode`でNode化したtoolから直接stateを更新(update)することはできず、`Command`を使う必要がある
- https://github.com/langchain-ai/langgraph/discussions/1322
- https://langchain-ai.github.io/langgraph/how-tos/tool-calling/

## 例
```python
@tool
def codebase_analysis(message: str, state: Annotated[State, InjectedState], tool_call_id: Annotated[str, InjectedToolCallId]) -> Command:
    """
    Analyze whether the (error) message/log is related to any code in the GitHub repository.
    Use this tool when the error or alert seems to be related to the codebase.

    Args:
        message: The (error) message/log to analyze.
        state: The state of the workflow.
        tool_call_id: The tool call id.
    Returns:
        The analysis result.
        Response format:
        {
            "analysis_results": {
                "response": {
                    "target github repository name": "The analysis result for the github repository."
                }
            }
        }
    """

    if "Failed to get github repositories" in state["git_repos"][0]:
        return "Github Repository is unavailable. Don't use this tool."

    lg.customlogger.info(f"[codebase_analysis] message: {message}")
    lg.customlogger.info(f"[codebase_analysis] git_repos: {state['git_repos']}")

    CODEBASE_ANALYSIS_ENDPOINT = os.getenv("CODEBASE_ANALYSIS_ENDPOINT")
    url = f"{CODEBASE_ANALYSIS_ENDPOINT}:8088/api/codebase-analysis"
    response = requests.post(url, json={"data": message, "git_repos": state["git_repos"]})
    lg.customlogger.info(f"[codebase_analysis] response: {response}")
    response_json = response.json()
    lg.customlogger.info(f"[codebase_analysis] response_json: {response_json}")
    
    ## CodebaseAnalysis側(Claude Code)でエラーメッセージと関係ないときは"false"だけ返すように指示しているが、余計な文章が混ざる場合があり、最後の5文字が"false"であるかを確認
    codebase_analysis_results = [{repo_name: analysis_result} for repo_name, analysis_result in response_json['analysis_results'].items() if analysis_result.lower().strip()[-5:] != "false" and analysis_result != "" and analysis_result is not None] # CodebaseAnalysisで分析結果、エラーと関係ない場合は"false"となるため、除外する
    lg.customlogger.info(f"[codebase_analysis] codebase_analysis_results: {codebase_analysis_results}")
    return Command(update={
        "codebase_analysis_results": codebase_analysis_results,
        "messages": [
            ToolMessage(f"Codebase analysis results: {codebase_analysis_results}", tool_call_id=tool_call_id)
        ]
    })
```