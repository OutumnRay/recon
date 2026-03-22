import numpy as np
import noisereduce as nr
from pydub import AudioSegment
from pathlib import Path


def enhance_audio(input_wav: Path, output_wav: Path):
    audio = AudioSegment.from_wav(input_wav)

    audio = audio.normalize()

    samples = np.array(audio.get_array_of_samples()).astype(np.float32)

    reduced_noise = nr.reduce_noise(
        y=samples,
        sr=16000,
        prop_decrease=0.8
    )

    enhanced = audio._spawn(
        reduced_noise.astype(np.int16).tobytes()
    )

    enhanced.export(output_wav, format="wav")
