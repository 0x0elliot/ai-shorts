import torch
from diffusers import StableDiffusionXLPipeline
from safetensors.torch import load_file

# Check if MPS is available
mps_available = hasattr(torch.backends, "mps") and torch.backends.mps.is_available()
device = torch.device("mps" if mps_available else "cpu")

# Path to your local safetensor file
local_safetensor_path = "/Users/aditya/Downloads/sd_xl_turbo_1.0_fp16.safetensors"

# Load the SDXL Turbo pipeline from local files
pipe = StableDiffusionXLPipeline.from_single_file(
    local_safetensor_path,
    torch_dtype=torch.float16,
    variant="fp16",
    use_safetensors=True
)

# Move the pipeline to MPS device
pipe = pipe.to(device)

# Define the prompt
prompt = "A serene landscape with mountains and a lake at sunset"

# Generate the image
image = pipe(
    prompt=prompt,
    num_inference_steps=1,  # SDXL Turbo is optimized for 1-4 steps
    guidance_scale=0.0,  # No classifier free guidance
).images[0]

# Save the generated image
image.save("generated_image.png")

print("Image generated and saved as 'generated_image.png'")