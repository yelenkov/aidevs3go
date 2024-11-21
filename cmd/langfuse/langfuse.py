import sys
import os
from langfuse.openai import openai
from dotenv import load_dotenv
import logging
import json

# Add the parent directory to the Python path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from config.models import Models
import requests

load_dotenv()

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

aidevs_api_key = os.getenv('AIDEVS_API_KEY')

def download_json():
    file_path = 'downloads/03.txt'
    
    if os.path.exists(file_path):
        logger.info("File already exists, skipping download.")
    else:
        url = f"https://centrala.ag3nts.org/data/{aidevs_api_key}/json.txt"
        response = requests.get(url)
        with open(file_path, 'wb') as file:
            file.write(response.content)

        logger.info("TXT file downloaded successfully.")

def validate_and_fix_equations():
    # Load the JSON data
    with open('downloads/03.txt', 'r') as file:
        data = json.load(file)
    
    corrections_made = 0
    test_data = data['test-data']
    
    for item in test_data:
        # Skip entries that have 'test' field
        if 'test' in item:
            continue
            
        # Parse the equation
        question = item['question']
        given_answer = item['answer']
        
        # Calculate the correct answer
        numbers = [int(x) for x in question.split(' + ')]
        correct_answer = sum(numbers)
        
        # Check if the answer is wrong and fix it
        if given_answer != correct_answer:
            item['answer'] = correct_answer
            corrections_made += 1
    
    # Save the corrected data back to file
    with open('downloads/03.txt', 'w') as file:
        json.dump(data, file, indent=4)
    
    return corrections_made

def handle_test_questions(test_data):
    # Initialize OpenAI client with API key
    openai.api_key = os.getenv('OPENAI_API_KEY')
    corrections_made = 0
    
    for item in test_data:
        # Only process items that have a 'test' field
        if 'test' in item:
            test_q = item['test']['q']
            
            # Get answer using OpenAI API
            response = openai.chat.completions.create(
                model=Models.GPT_4O_MINI.full_name,  # Use the model string from models.py
                messages=[
                    {"role": "user", "content": f"Please answer this question concisely: {test_q}"}
                ],
                temperature=0.2  # Set temperature to 0.2
            )
            
            answer = response.choices[0].message.content.strip()
            
            # Update the answer in the data
            item['test']['a'] = answer
            corrections_made += 1
            
            logger.info(f"Question: {test_q}")
            logger.info(f"Answer: {answer}")
            logger.info("---")
    
    return corrections_made

def process_file():
    # Load the JSON data
    with open('downloads/03.txt', 'r') as file:
        data = json.load(file)
    
    # First fix the math equations
    math_corrections = validate_and_fix_equations()
    
    # Then handle the test questions
    test_corrections = handle_test_questions(data['test-data'])
    
    # Save all corrections back to file
    with open('downloads/03.txt', 'w') as file:
        json.dump(data, file, indent=4)
    
    return math_corrections, test_corrections

def send_report():
    # Read the processed file
    with open('downloads/03.txt', 'r') as file:
        processed_data = json.load(file)
    
    # Prepare the payload
    payload = {
        "task": "JSON",
        "apikey": aidevs_api_key,
        "answer": processed_data
    }
    
    # Send POST request to report endpoint
    response = requests.post('https://centrala.ag3nts.org/report', json=payload)
    
    # Log the response
    if response.status_code == 200:
        logger.info(f"Report sent successfully: {response.json()}")
    else:
        logger.error(f"Error sending report: {response.status_code} - {response.text}")
    
    return response.json()

if __name__ == "__main__":
    download_json()
    
    # Fix both math answers and test questions
    math_fixes, test_fixes = process_file()
    
    # Print results
    if math_fixes > 0:
        logger.info(f"Fixed {math_fixes} incorrect math answers in the file.")
    else:
        logger.info("All math equations were correct!")
        
    if test_fixes > 0:
        logger.info(f"Answered {test_fixes} test questions in the file.")
    
    # Send the report
    report_response = send_report()