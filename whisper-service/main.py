import tempfile
from faster_whisper import WhisperModel
from flask import Flask, request, jsonify

app = Flask(__name__)

def generate_asr_data(audio_file):
    model = WhisperModel("base", device="cpu", compute_type="int8")
    segments, _ = model.transcribe(audio_file, word_timestamps=True)

    asr_data = {
        "sentences": [],
        "words": []
    }

    for segment in segments:
        asr_data["sentences"].append({
            "start": segment.start,
            "end": segment.end,
            "text": segment.text
        })
        for word in segment.words:
            asr_data["words"].append({
                "start": word.start,
                "end": word.end,
                "word": word.word
            })

    return asr_data

@app.route('/generate_asr', methods=['POST'])
def generate_asr():
    if 'audio' not in request.files:
        return jsonify({"error": "No audio file provided"}), 400
    
    audio_file = request.files['audio']
    
    if audio_file.filename == '':
        return jsonify({"error": "No selected file"}), 400
    
    if audio_file and audio_file.filename.lower().endswith(('.mp3', '.wav', '.flac')):
        with tempfile.NamedTemporaryFile(delete=False, suffix="." + audio_file.filename.split('.')[-1]) as temp_audio:
            audio_file.save(temp_audio.name)
            asr_data = generate_asr_data(temp_audio.name)
        
        return jsonify(asr_data)
    else:
        return jsonify({"error": "Invalid file format. Please upload an MP3, WAV, or FLAC file."}), 400

if __name__ == "__main__":
    app.run(debug=True)