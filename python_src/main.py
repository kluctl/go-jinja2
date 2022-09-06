import json
import sys

from go_jinja2.jinja2_renderer import Jinja2Renderer


def main():
    while True:
        args = sys.stdin.readline()
        if not args:
            break
        args = json.loads(args)
        opts = args["opts"]

        r = Jinja2Renderer(opts)

        if args["cmd"] == "render-strings":
            result = r.RenderStrings(args["templates"])
        elif args["cmd"] == "render-files":
            result = r.RenderFiles(args["templates"])
        elif args["cmd"] == "exit":
            break
        else:
            raise Exception("invalid cmd")

        result = json.dumps(result)

        sys.stdout.write(result + "\n")
        sys.stdout.flush()

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt as e:
        pass
