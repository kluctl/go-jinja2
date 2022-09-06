from jinja2 import StrictUndefined, FileSystemLoader, ChainableUndefined

from .jinja2_utils import MyEnvironment, extract_template_error


class NullUndefined(ChainableUndefined):
    def _return_self(self, other):
        return self

    __add__ = __radd__ = __sub__ = __rsub__ = _return_self
    __mul__ = __rmul__ = __div__ = __rdiv__ = _return_self
    __truediv__ = __rtruediv__ = _return_self
    __floordiv__ = __rfloordiv__ = _return_self
    __mod__ = __rmod__ = _return_self
    __pos__ = __neg__ = _return_self
    __call__ = __getitem__ = _return_self
    __lt__ = __le__ = __gt__ = __ge__ = _return_self
    __int__ = __float__ = __complex__ = _return_self
    __pow__ = __rpow__ = _return_self

class Jinja2Renderer:
    def __init__(self, opts):
        self.opts = opts

    def build_env(self):
        environment = MyEnvironment(loader=FileSystemLoader(self.opts.get("searchDirs", [])),
                                    undefined=StrictUndefined if self.opts.get("strict", True) else NullUndefined,
                                    cache_size=10000,
                                    auto_reload=False,
                                    trim_blocks=self.opts.get("trimBlocks", False),
                                    lstrip_blocks=self.opts.get("lstripBlocks", False))
        environment.globals.update(self.opts.get("globals", {}))

        for e in self.opts.get("extensions", []):
            environment.add_extension(e)

        return environment

    def render_helper(self, templates, is_string):
        env = self.build_env()

        result = []

        for i, t in enumerate(templates):
            try:
                if is_string:
                    t = env.from_string(t)
                else:
                    t = env.get_template(t)
                result.append({
                    "result": t.render()
                })
            except Exception as e:
                result.append({
                    "error": extract_template_error(e),
                })

        return result

    def RenderStrings(self, templates):
        try:
            return self.render_helper(templates, True)
        except Exception as e:
            return [{
                "error": str(e)
            }] * len(templates)

    def RenderFiles(self, templates):
        try:
            return self.render_helper(templates, False)
        except Exception as e:
            return [{
                "error": str(e)
            }] * len(templates)
