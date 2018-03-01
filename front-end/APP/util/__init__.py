#! /usr/bin/env python
# -*- coding: utf-8 -*-
from flask import Blueprint
filter_blueprint = Blueprint('filters', __name__)

# Register all the filter.
from . import file_manage, API_manage, filters
