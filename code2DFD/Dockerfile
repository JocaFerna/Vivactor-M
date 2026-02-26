# Use a lightweight Python base image
FROM python:3.11-slim

# Set the working directory inside the container
WORKDIR /app

# Install system dependencies (needed for git or potential C-extensions)
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first to leverage Docker cache
# (Assumes you have a requirements.txt in the same folder)
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code
COPY . .

# Expose the port the app runs on (matching your compose and script default)
EXPOSE 5001

# Set environment variables
ENV FLASK_APP=flask_code2DFD.py
ENV PORT=5001

# Run the application
CMD ["python","-u", "flask_code2DFD.py"]