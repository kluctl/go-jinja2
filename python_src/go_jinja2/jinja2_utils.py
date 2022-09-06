import io
import os
import sys
import traceback

from jinja2 import Environment, TemplateNotFound


# This Jinja2 environment allows to load templates relative to the parent template. This means that for example
# '{% include "file.yml" %}' will try to include the template from a ./file.yml
class MyEnvironment(Environment):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.add_extension("jinja2.ext.loopcontrols")

    """Override join_path() to enable relative template paths."""
    """See https://stackoverflow.com/a/3655911/7132642"""
    def join_path(self, template, parent):
        p = os.path.join(os.path.dirname(parent), template)
        p = os.path.normpath(p)
        return p.replace('\\', '/')

def extract_template_error(e):
    try:
        raise e
    except TemplateNotFound as e2:
        return "template %s not found" % str(e2)
    except:
        etype, value, tb = sys.exc_info()
    extracted_tb = traceback.extract_tb(tb)
    found_template = None
    for i, s in reversed(list(enumerate(extracted_tb))):
        if not s.filename.endswith(".py"):
            found_template = i
            break
    f = io.StringIO()
    if found_template is not None:
        traceback.print_list([extracted_tb[found_template]], file=f)
        print("%s: %s" % (type(e).__name__, str(e)), file=f)
    else:
        traceback.print_exception(etype, value, tb, file=f)
    return f.getvalue()
