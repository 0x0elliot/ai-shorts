import tempfile
import Levenshtein
from faster_whisper import WhisperModel
from flask import Flask, request, jsonify

app = Flask(__name__)

def generate_asr_data(audio_file, original_script):
    model = WhisperModel("base", device="cpu", compute_type="int8")
    segments, _ = model.transcribe(audio_file, word_timestamps=True)
    
    # Preprocess the original script
    original_words = original_script.lower().split()
    
    asr_data = {
        "sentences": [],
        "words": []
    }
    
    word_index = 0
    for segment in segments:
        corrected_segment = []
        for word in segment.words:
            if word_index < len(original_words):
                # Find the closest matching word from the original script
                # closest_word = min(original_words[max(0, word_index-1):min(len(original_words), word_index+1)], 
                #                    key=lambda x: Levenshtein.distance(word.word, x))
                
                # assume that the word from the script is the closest match
                closest_word = original_words[word_index]
                
                asr_data["words"].append({
                    "start": word.start,
                    "end": word.end,
                    "word": closest_word,
                    # "original_word": word.word
                })
                corrected_segment.append(closest_word)
                word_index += 1
            else:
                # If we've run out of words in the original script, use the ASR word
                asr_data["words"].append({
                    "start": word.start,
                    "end": word.end,
                    "word": word.word,
                    # "original_word": word.word
                })
                corrected_segment.append(word.word)
        
        corrected_text = " ".join(corrected_segment)
        asr_data["sentences"].append({
            "start": segment.start,
            "end": segment.end,
            "text": corrected_text,
            # "original_text": segment.text
        })
    
    return asr_data

@app.route('/generate_asr', methods=['POST'])
def generate_asr():
    if 'audio' not in request.files:
        return jsonify({"error": "No audio file provided"}), 400
    
    audio_file = request.files['audio']
    original_script = request.form.get('original_script', None)
    
    if audio_file.filename == '':
        return jsonify({"error": "No selected file"}), 400
    
    if audio_file and audio_file.filename.lower().endswith(('.mp3', '.wav', '.flac')):
        with tempfile.NamedTemporaryFile(delete=False, suffix="." + audio_file.filename.split('.')[-1]) as temp_audio:
            audio_file.save(temp_audio.name)
        asr_data = generate_asr_data(temp_audio.name, original_script)
        
        return jsonify(asr_data)
    else:
        return jsonify({"error": "Invalid file format. Please upload an MP3, WAV, or FLAC file."}), 400

if __name__ == "__main__":
    app.run(debug=True)