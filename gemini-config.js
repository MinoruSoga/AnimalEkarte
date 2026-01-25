/**
 * Gemini API Configuration Example
 * Based on official Google best practices for 2025
 */

const { GoogleGenerativeAI } = require("@google/generative-ai");
require("dotenv").config();

class GeminiConfig {
  constructor(apiKey = null) {
    this.apiKey = apiKey || process.env.GEMINI_API_KEY;
    if (!this.apiKey) {
      throw new Error(
        "API key must be provided or set in GEMINI_API_KEY environment variable"
      );
    }
    this.genAI = new GoogleGenerativeAI(this.apiKey);
  }

  /**
   * Configuration for complex reasoning tasks
   */
  getComplexTaskConfig() {
    return {
      generationConfig: {
        temperature: 1.0, // Default for Gemini 3
        maxOutputTokens: 8192,
        thinkingConfig: {
          thinkingLevel: "HIGH", // Maximum reasoning for complex tasks
        },
      },
    };
  }

  /**
   * Configuration for simple, fast tasks
   */
  getSimpleTaskConfig() {
    return {
      generationConfig: {
        temperature: 0.1, // Low temperature for consistency
        maxOutputTokens: 1024,
        thinkingConfig: {
          thinkingLevel: "LOW", // Minimal reasoning for speed
        },
      },
    };
  }

  /**
   * Configuration optimized for code generation
   */
  getCodingTaskConfig() {
    return {
      generationConfig: {
        temperature: 0.7, // Balanced creativity and consistency
        maxOutputTokens: 4096,
        thinkingConfig: {
          thinkingLevel: "HIGH", // Complex reasoning for coding
        },
      },
    };
  }

  /**
   * Generate response for complex tasks with high thinking level
   */
  async generateComplexResponse(prompt, model = "gemini-3-pro-preview") {
    const modelInstance = this.genAI.getGenerativeModel({
      model,
      ...this.getComplexTaskConfig(),
    });

    const result = await modelInstance.generateContent(prompt);
    return result.response.text();
  }

  /**
   * Generate response for simple tasks with low thinking level
   */
  async generateSimpleResponse(prompt, model = "gemini-3-pro-preview") {
    const modelInstance = this.genAI.getGenerativeModel({
      model,
      ...this.getSimpleTaskConfig(),
    });

    const result = await modelInstance.generateContent(prompt);
    return result.response.text();
  }

  /**
   * Generate code with optimized configuration
   */
  async generateCode(prompt, model = "gemini-3-pro-preview") {
    const modelInstance = this.genAI.getGenerativeModel({
      model,
      ...this.getCodingTaskConfig(),
    });

    const result = await modelInstance.generateContent(prompt);
    return result.response.text();
  }
}

/**
 * Example usage of GeminiConfig
 */
async function main() {
  try {
    // Initialize configuration
    const gemini = new GeminiConfig();

    // Complex reasoning example
    const complexPrompt = `
    Analyze this business problem and provide a step-by-step solution:
    
    Problem: A company wants to reduce customer churn by 25% in the next quarter.
    Current churn rate is 15%. The company has 10,000 customers.
    
    Based on the information above, provide a comprehensive strategy with specific metrics and timelines.
    `;

    const complexResult = await gemini.generateComplexResponse(complexPrompt);
    console.log("Complex Analysis Result:");
    console.log(complexResult);

    // Simple task example
    const simplePrompt =
      "Summarize this text in one sentence: Machine learning is a subset of artificial intelligence that enables systems to learn and improve from experience without being explicitly programmed.";

    const simpleResult = await gemini.generateSimpleResponse(simplePrompt);
    console.log("\nSimple Summary Result:");
    console.log(simpleResult);

    // Code generation example
    const codePrompt = `
    Write a JavaScript function that:
    1. Takes an array of objects as input
    2. Each object has 'name' and 'score' properties
    3. Returns the top 3 performers by score
    4. Includes proper error handling and JSDoc comments
    `;

    const codeResult = await gemini.generateCode(codePrompt);
    console.log("\nGenerated Code:");
    console.log(codeResult);
  } catch (error) {
    console.error("Error:", error.message);
  }
}

module.exports = GeminiConfig;

// Run example if this file is executed directly
if (require.main === module) {
  main();
}
