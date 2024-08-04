import sys
import os
import pysrt
from moviepy.editor import *

def time_to_seconds(time_obj):
    return time_obj.hours * 3600 + time_obj.minutes * 60 + time_obj.seconds + time_obj.milliseconds / 1000

def create_ken_burns_clip(img_path, duration):
    img = ImageClip(img_path).set_duration(duration)
    w, h = img.size
    crop_factor = 1.1

    def ken_burns(t):
        zoom = 1 + t/duration * (crop_factor-1)
        new_w = int(w * zoom)
        new_h = int(h * zoom)
        resized = img.resize((new_w, new_h))
        return resized.crop(x1=int(new_w/2-w/2), y1=int(new_h/2-h/2), x2=int(new_w/2+w/2), y2=int(new_h/2+h/2)).get_frame(t)
    
    return VideoClip(make_frame=ken_burns, duration=duration)

def create_slideshow_with_subtitles(image_paths, svt_file, audio_file, output_file):
    # Load SVT file
    subs = pysrt.open(svt_file)

    # Create clips with Ken Burns effect
    clips = []
    for i, sub in enumerate(subs):
        if i >= len(image_paths):
            break
        duration = time_to_seconds(sub.end) - time_to_seconds(sub.start)
        clip = create_ken_burns_clip(image_paths[i], duration)
        clips.append(clip)

    # Concatenate image clips
    video = concatenate_videoclips(clips)

    def create_subtitle_clips(subs, video_size):
        w, h = video_size
        return [((time_to_seconds(sub.start), time_to_seconds(sub.end)),
                TextClip(sub.text, 
                        font='Roboto-Bold', 
                        fontsize=h//15,  # Adjust this value to make text bigger or smaller
                        color='white',
                        stroke_color='black',
                        stroke_width=2,
                        method='caption',
                        size=(w*0.8, None),  # Wrap text at 80% of video width
                        align='center')
                .set_position(('center', 'center')))
                for sub in subs[:len(clips)]]
        
    # Get video size from the first image
    first_img = ImageClip(image_paths[0])
    video_size = first_img.size

    subtitles = create_subtitle_clips(subs, video_size)
    subtitles_clip = CompositeVideoClip([video] + [clip.set_start(t[0]).set_end(t[1]) for t, clip in subtitles])

    # Load audio
    audio = AudioFileClip(audio_file)

    # Set video duration to match audio duration if necessary
    if subtitles_clip.duration > audio.duration:
        subtitles_clip = subtitles_clip.set_duration(audio.duration)
    else:
        audio = audio.subclip(0, subtitles_clip.duration)

    # Add audio to video
    final_video = subtitles_clip.set_audio(audio)

    # Write output file
    final_video.write_videofile(output_file, fps=30, audio_codec='aac')

if __name__ == "__main__":
    # assume everything is in the /tmp directory
    # get all files in /tmp that start with image_
    image_paths = sorted([f'/tmp/{f}' for f in os.listdir('/tmp') if f.startswith('image_')])
    srt_file = '/tmp/subtitles.srt'
    audio_file = '/tmp/full_audio.mp3'
    output_file = '/tmp/output.mp4'

    create_slideshow_with_subtitles(image_paths, srt_file, audio_file, output_file)