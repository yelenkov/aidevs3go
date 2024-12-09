from dotenv import load_dotenv
from langfuse import Langfuse

load_dotenv()

langfuse = Langfuse()

print(langfuse.auth_check())

trace = langfuse.trace(name = "llm-feature")
retrieval = trace.span(name = "retrieval")
retrieval.generation(name = "query-creation")
retrieval.span(name = "vector-db-search")
retrieval.event(name = "db-summary")


# creates generation
generation = trace.generation(
    name="summary-generation",
    model="gpt-4o-mini",
    model_parameters={"maxTokens": "1000", "temperature": "0.9"},
    input=[{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "Please generate a summary of the following documents \nThe engineering department defined the following OKR goals...\nThe marketing department defined the following OKR goals..."}],
    metadata={"interface": "whatsapp"}
)
 
# execute model, mocked here
# chat_completion = openai.ChatCompletion.create(model="gpt-3.5-turbo", messages=[{"role": "user", "content": "Hello world"}])
chat_completion = {
    "completion":"The Q3 OKRs contain goals for multiple teams...",
    "usage":{"input": 50, "output": 49, "unit":"TOKENS"}
}
 
# update span and sets end_time
generation.end(
    output=chat_completion["completion"],
    usage=chat_completion["usage"],
);