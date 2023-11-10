import os, datetime
import sys, subprocess, jinja2


def download_dataset(count, to):
    FILE_PER_CURL = 50
    i = 0
    while count > 0:
        request_for = min(count, FILE_PER_CURL)
        subprocess.run(
            [
                "curl",
                f"https://unsample.net/archive?count={request_for}&width=1920&height=1080",
                "--output",
                f"{to}/a{i}.zip",
            ])
        subprocess.run(
            ["unzip", "-d", f"{to}/a{i}", f"{to}/a{i}.zip"],
        )
        count -= FILE_PER_CURL
        i += 1


def get_paths(root_dir):
    output = []
    files = os.listdir(root_dir)
    for file in files:
        if os.path.isfile(root_dir + "/" + file):
            output += [f"{root_dir}/{file}"]
        else:
            output += get_paths(root_dir + "/" + file)
    return output


args = sys.argv[1:]
media_dir = args[args.index("-media-dir") + 1]
download_data = "-init-dataset" in args

if download_data:
    download_dataset(int(args[args.index("-init-dataset") + 1]), media_dir)

templateLoader = jinja2.FileSystemLoader(searchpath=".")
templateEnv = jinja2.Environment(loader=templateLoader)

TEMPLATE_FILE = "template.html"
template = templateEnv.get_template(TEMPLATE_FILE)

pictures = get_paths(media_dir)
if "-prefix" in args:
    prefix = args[args.index("-prefix") + 1]
    pictures = [
        {"key": picture.replace(media_dir + "/", prefix), "last_modified": datetime.datetime.now() - datetime.timedelta(days=30)}
        for picture in pictures
    ]
outputText = template.render({"pictures": pictures})

with open("index.html", "w") as f:
    f.write(outputText)

subprocess.run([
    "mv",
    "index.html",
    media_dir
])
