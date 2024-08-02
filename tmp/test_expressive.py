from TTS.api import TTS
import torch

# Check if MPS is available
if torch.backends.mps.is_available():
    device = "mps"
    print("Using MPS (Apple Silicon GPU)")
else:
    device = "cpu"
    print("MPS not available, using CPU")

# Initialize TTS with the appropriate device
tts = TTS("tts_models/en/multi-dataset/tortoise-v2").to(device)

# Get all speakers
print(tts.speakers)

# Generate audio files
tts.tts_to_file(text="Hello, my name is Manmay, how are you?",
                file_path="output1.wav",
                voice_dir="/Users/aditya/Documents/OSS/zappush/shortpro/tmp/",
                num_autoregressive_samples=1,
                diffusion_iterations=10)

tts.tts_to_file(text="Hello, my name is Manmay, how are you?",
                file_path="output2.wav",
                voice_dir="/Users/aditya/Documents/OSS/zappush/shortpro/tmp/",
                preset="ultra_fast")

tts.tts_to_file(text="Hello, my name is Manmay, how are you?",
                file_path="output3.wav")