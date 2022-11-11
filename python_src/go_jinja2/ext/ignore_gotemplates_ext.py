import re
import typing

from jinja2.ext import Extension
from jinja2.lexer import Token

_gotemplate_start_re = re.compile(r"\{\{([ \t]*)\.")

dummy = "dummyignoregotemplatesreplacementmarker"


class IgnoreGoTemplates(Extension):

    def preprocess(
        self, source: str, name: typing.Optional[str], filename: typing.Optional[str] = None
    ) -> str:
        source = _gotemplate_start_re.sub(r"%s\1." % dummy, source)
        return source

    def filter_stream(self, stream):
        for token in stream:
            if token.type in ["data", "name"] and dummy in token.value:
                yield Token(token.lineno, "data", token.value.replace(dummy, "{{"))
            elif token.type == "string":
                yield Token(token.lineno, "string", token.value.replace(dummy, "{{"))
            else:
                yield token
