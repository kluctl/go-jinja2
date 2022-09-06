import base64
import hashlib
import json
import os
import sys

import jinja2
from jinja2 import TemplateError, TemplateNotFound
from jinja2.ext import Extension
from jinja2.runtime import Context

from .dict_utils import get_dict_value, merge_dict
from .yaml_utils import yaml_dump, yaml_load


class KluctlExtension(Extension):
    def __init__(self, environment):
        super().__init__(environment)
        add_jinja2_filters(environment)


def b64encode(string):
    return base64.b64encode(string.encode()).decode()


def b64decode(string):
    return base64.b64decode(string.encode()).decode()


def to_yaml(obj):
    return yaml_dump(obj)


def to_json(obj):
    return json.dumps(obj)


def from_yaml(s):
    return yaml_load(s)


@jinja2.pass_context
def render(ctx, string):
    t = ctx.environment.from_string(string)
    return t.render(ctx.parent)


def sha256(s):
    if not isinstance(s, bytes):
        s = s.encode("utf-8")
    return hashlib.sha256(s).hexdigest()


@jinja2.pass_context
def load_template(ctx, path, **kwargs):
    dir = os.path.dirname(ctx.name)
    full_path = os.path.join(dir, path)
    try:
        t = ctx.environment.get_template(full_path.replace('\\', '/'))
    except TemplateNotFound:
        t = ctx.environment.get_template(path.replace('\\', '/'))
    vars = merge_dict(ctx.parent, kwargs)
    return t.render(vars)


class VarNotFoundException(Exception):
    pass


@jinja2.pass_context
def get_var(ctx, path, default):
    if not isinstance(path, list):
        path = [path]
    for p in path:
        r = get_dict_value(ctx.parent, p, VarNotFoundException())
        if isinstance(r, VarNotFoundException):
            continue
        return r
    return default


def update_dict(a, b):
    merge_dict(a, b, False)
    return ""


def raise_helper(msg):
    raise TemplateError(msg)


def debug_print(msg):
    sys.stderr.write("debug_print: %s\n" % str(msg))
    return ""


@jinja2.pass_context
def load_sha256(ctx: Context, path, digest_len=None):
    if "__calc_sha256__" in ctx:
        return "__self_sha256__"
    rendered = load_template(ctx, path, __calc_sha256__=True)
    hash = hashlib.sha256(rendered.encode("utf-8")).hexdigest()
    if digest_len is not None:
        hash = hash[:digest_len]
    return hash


def add_jinja2_filters(jinja2_env):
    jinja2_env.filters['b64encode'] = b64encode
    jinja2_env.filters['b64decode'] = b64decode
    jinja2_env.filters['to_yaml'] = to_yaml
    jinja2_env.filters['to_json'] = to_json
    jinja2_env.filters['from_yaml'] = from_yaml
    jinja2_env.filters['render'] = render
    jinja2_env.filters['sha256'] = sha256
    jinja2_env.globals['load_template'] = load_template
    jinja2_env.globals['get_var'] = get_var
    jinja2_env.globals['merge_dict'] = merge_dict
    jinja2_env.globals['update_dict'] = update_dict
    jinja2_env.globals['raise'] = raise_helper
    jinja2_env.globals['debug_print'] = debug_print
    jinja2_env.globals['load_sha256'] = load_sha256
