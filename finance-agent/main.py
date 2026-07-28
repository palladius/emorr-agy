from google_antigravity import Agent, LocalAgentConfig

def main():
    # Scaffolded Finance Agent
    config = LocalAgentConfig()
    agent = Agent(
        config=config,
        system_instruction="You are a finance agent designed to help with financial analysis and reporting."
    )
    print("Finance Agent initialized successfully!")

if __name__ == "__main__":
    main()
