from flask import Flask, request, jsonify
from openai import OpenAI
from langfuse.decorators import observe, langfuse_context
from dotenv import load_dotenv
import os
load_dotenv()

app = Flask(__name__)

client = OpenAI(
    base_url="https://openrouter.ai/api/v1",
    api_key=os.getenv("OPENROUTER_API_KEY"),
)

@observe(as_type="generation")
def openrouter_completion(**kwargs):
    kwargs_clone = kwargs.copy()
    input = kwargs_clone.pop('messages', None)
    model = kwargs_clone.pop('model', None)
    langfuse_context.update_current_observation(
        input=input,
        model="claude-3-5-haiku-20241022",
        metadata=kwargs_clone
    )

    response = client.chat.completions.create(**kwargs)
  
    # See docs for more details on token counts and usd cost in Langfuse
    # https://langfuse.com/docs/model-usage-and-cost
    # langfuse_context.update_current_observation(
    #     usage={
    #       "input": response.usage.input_tokens,
    #       "output": response.usage.output_tokens
    #   }
#   )
 
  # return result
    print(response.choices[0].message.content)
    return response.choices[0].message.content

@observe()
def main():
    return openrouter_completion(
        model="anthropic/claude-3-5-haiku-20241022",
        messages=[
            {
            "role": "user",
            "content": "What is the meaning of death?"
        }
        ]
    )
main()