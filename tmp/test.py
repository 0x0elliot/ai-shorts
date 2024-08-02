from TTS.api import TTS

print("Initializing TTS model...")
# Initialize TTS with a model that has a deeper male voice
tts = TTS(model_name="tts_models/en/vctk/vits")

# Text to be converted to speech
text = "Hey! You can use my voice to leave your shorts on auto pilot!"

print(f"Generating speech for text: '{text}'")

fav_speakers = {
    "Edwards": "p230",
    "Elena": "p248",
    "Oliver": "p251",
    "James": "p254",
    "William": "p256",
    "Charlotte": "p260",
    "Sophia": "p263",
    "Michael": "p264",
    "Daniel": "p267",
    "Emily": "p273",
    "Thomas": "p282",
    "Priya": "p345"
}

for speaker, speaker_id in fav_speakers.items():
    # remember, we need to make the speaker sound a little more natural and excited
    # it's an instagram short after all
    audio = tts.tts_to_file(text=text, speaker=speaker_id, file_path=f"{speaker}.wav", speed=1.2)

    print(f"Audio saved to tmp/{speaker}.wav")