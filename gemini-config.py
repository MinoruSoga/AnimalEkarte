"""
Gemini API Configuration Example
Based on official Google best practices for 2025
"""

from google import genai
from google.genai import types
import os
from typing import Optional


class GeminiConfig:
    """Configuration class for Gemini API with best practices"""
    
    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv('GEMINI_API_KEY')
        if not self.api_key:
            raise ValueError("API key must be provided or set in GEMINI_API_KEY environment variable")
        
        self.client = genai.Client(api_key=self.api_key)

    def get_config_for_complex_task(self) -> types.GenerateContentConfig:
        """Configuration for complex reasoning tasks"""
        return types.GenerateContentConfig(
            thinking_config=types.ThinkingConfig(
                thinking_level="HIGH"  # Maximum reasoning for complex tasks
            ),
            temperature=1.0,  # Default for Gemini 3
            max_output_tokens=8192,
        )

    def get_config_for_simple_task(self) -> types.GenerateContentConfig:
        """Configuration for simple, fast tasks"""
        return types.GenerateContentConfig(
            thinking_config=types.ThinkingConfig(
                thinking_level="LOW"  # Minimal reasoning for speed
            ),
            temperature=0.1,  # Low temperature for consistency
            max_output_tokens=1024,
        )

    def get_config_for_coding_task(self) -> types.GenerateContentConfig:
        """Configuration optimized for code generation"""
        return types.GenerateContentConfig(
            thinking_config=types.ThinkingConfig(
                thinking_level="HIGH"  # Complex reasoning for coding
            ),
            temperature=0.7,  # Balanced creativity and consistency
            max_output_tokens=4096,
        )

    async def generate_complex_response(
        self, prompt: str, model: str = "gemini-3-pro-preview"
    ) -> str:
        """Generate response for complex tasks with high thinking level"""
        config = self.get_config_for_complex_task()
        response = self.client.models.generate_content(
            model=model,
            contents=prompt,
            config=config
        )
        return response.text

    async def generate_simple_response(
        self, prompt: str, model: str = "gemini-3-pro-preview"
    ) -> str:
        """Generate response for simple tasks with low thinking level"""
        config = self.get_config_for_simple_task()
        response = self.client.models.generate_content(
            model=model,
            contents=prompt,
            config=config
        )
        return response.text

    async def generate_code(
        self, prompt: str, model: str = "gemini-3-pro-preview"
    ) -> str:
        """Generate code with optimized configuration"""
        config = self.get_config_for_coding_task()
        response = self.client.models.generate_content(
            model=model,
            contents=prompt,
            config=config
        )
        return response.text


# Usage examples
async def main():
    """Example usage of GeminiConfig"""
    
    # Initialize configuration
    gemini = GeminiConfig()

    # Complex reasoning example
    complex_prompt = """
    Analyze this business problem and provide a step-by-step solution:
    
    Problem: A company wants to reduce customer churn by 25% in the next quarter.
    Current churn rate is 15%. The company has 10,000 customers.
    
    Based on the information above, provide a comprehensive strategy with specific metrics and timelines.
    """
    
    complex_result = await gemini.generate_complex_response(complex_prompt)
    print("Complex Analysis Result:")
    print(complex_result)

    # Simple task example
    simple_prompt = "Summarize this text in one sentence: Machine learning is a subset of artificial intelligence that enables systems to learn and improve from experience without being explicitly programmed."
    
    simple_result = await gemini.generate_simple_response(simple_prompt)
    print("\nSimple Summary Result:")
    print(simple_result)

    # Code generation example
    code_prompt = """
    Write a Python function that:
    1. Takes a list of dictionaries as input
    2. Each dictionary has 'name' and 'score' keys
    3. Returns the top 3 performers by score
    # 4. Includes proper error handling and type hints
    """
    
    code_result = await gemini.generate_code(code_prompt)
    print("\nGenerated Code:")
    print(code_result)


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
