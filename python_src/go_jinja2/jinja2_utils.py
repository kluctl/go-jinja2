import io
import os
import sys
import traceback
import typing

from jinja2 import Environment, TemplateNotFound, BaseLoader


# This Jinja2 environment allows to load templates relative to the parent template. This means that for example
# '{% include "file.yml" %}' will try to include the template from a ./file.yml
class MyEnvironment(Environment):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    """Override join_path() to enable relative template paths."""
    """See https://stackoverflow.com/a/3655911/7132642"""
    def join_path(self, template, parent):
        if template[:2] == "./":
            p = os.path.join(os.path.dirname(parent), template)
            p = os.path.normpath(p)
            return p.replace('\\', '/')
        return template


def _read_template_helper(template):
    try:
        with open(template) as f:
            contents = f.read()
    except OSError:
        raise TemplateNotFound(template)
    mtime = os.path.getmtime(template)

    def uptodate() -> bool:
        try:
            return os.path.getmtime(template) == mtime
        except OSError:
            return False

    return contents, os.path.normpath(template), uptodate

class RootTemplateLoader(BaseLoader):
    def __init__(self):
        super().__init__()
        self.root_template = None

    def get_source(
            self, environment: "Environment", template: str
    ) -> typing.Tuple[str, typing.Optional[str], typing.Optional[typing.Callable[[], bool]]]:
        if template != self.root_template:
            raise TemplateNotFound(template)
        return _read_template_helper(template)


class SearchPathAbsLoader(BaseLoader):
    def __init__(self, searchpath):
        self.searchpath = searchpath

    def get_source(
            self, environment: "Environment", template: str
    ) -> typing.Tuple[str, str, typing.Callable[[], bool]]:
        try:
            if not os.path.isabs(template):
                raise TemplateNotFound(template)
            if os.path.abspath(template) != template:
                raise TemplateNotFound(template)
            found = False
            for s in self.searchpath:
                sabs = os.path.abspath(s)
                if template.startswith(sabs):
                    found = True
            if not found:
                raise TemplateNotFound(template)
        except OSError:
            raise TemplateNotFound(template)



        return _read_template_helper(template)


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
