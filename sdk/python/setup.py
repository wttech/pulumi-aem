# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import errno
import os
from setuptools import setup, find_packages
from setuptools.command.install import install
from subprocess import check_call


VERSION = os.getenv("PULUMI_PYTHON_VERSION", "0.0.0")
def readme():
    try:
        with open('README.md', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "aem Pulumi Package - Development Version"


setup(name='wttech_aem',
      python_requires='>=3.8',
      version=VERSION,
      description="Easily manage AEM instances in the cloud without a deep dev-ops knowledge",
      long_description=readme(),
      long_description_content_type='text/markdown',
      keywords='pulumi aem aemc cloud',
      url='https://github.com/wttech/pulumi-aem',
      project_urls={
          'Repository': 'https://github.com/wttech/pulumi-aem'
      },
      license='Apache-2.0',
      packages=find_packages(),
      package_data={
          'wttech_aem': [
              'py.typed',
              'pulumi-plugin.json',
          ]
      },
      install_requires=[
          'parver>=0.2.1',
          'pulumi>=3.56.0,<4.0.0',
          'semver>=2.8.1'
      ],
      zip_safe=False)
