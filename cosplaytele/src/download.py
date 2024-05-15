import os
import requests
import shutil

def download_file(url: str) -> str:
    local_filename = url.split('/')[-1]
    print("local_filename: ", local_filename)

    with requests.get(url, stream=True) as r:
        r.raise_for_status()

        with open(local_filename, 'wb') as f:
            shutil.copyfileobj(r.raw, f)

    return local_filename
