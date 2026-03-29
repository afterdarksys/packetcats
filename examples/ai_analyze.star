# ai_analyze.star

# Using the Pipeline Decode feature
def packet_hook(raw):
    # Pass the raw payload to the AI to check for malicious intent
    # Provide explicitly 'anthropic', 'openai', or 'gemini'
    analysis = ai.analyze("Does this network payload contain a SQL injection attempt or directory traversal?", raw, provider="openai")
    
    print("--- AI Analysis ---")
    print(analysis)
    print("-------------------\n")
