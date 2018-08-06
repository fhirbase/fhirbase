import json
import logging
import zipfile
import io


try:
    import requests
except ImportError:
    raise RuntimeError('Run `pip install requests` before using this tool')


logger = logging.getLogger()


def isndjson(filename):
    return filename.endswith('.ndjson')


def isjson(filename):
    return filename.endswith('.json')


def iszip(filename):
    return filename.endswith('.zip')


def dumpjson(res):
    return json.dumps(res, separators=(',', ':'))


def loadjson(jsonstr):
    return json.loads(jsonstr)


def prepare_json(jsonstr):
    """
    Processes Bundle or another resource and
    returns its ndjson representation as list
    """
    res = loadjson(jsonstr)

    resource_type = res.get('resourceType', None)
    if not resource_type:
        return []

    if resource_type == 'Bundle' and 'entry' in res:
        return [dumpjson(item['resource'])
                for item in res['entry']
                if 'resource' in item]
    else:
        return [dumpjson(res)]


def make_zip_openfn(zipfile):
    """
    Zipfile only opens file in binary mode but we want to work with text
    See https://stackoverflow.com/q/5627954
    """
    return lambda name, mode: io.TextIOWrapper(zipfile.open(name, mode))


def iter_lines_from_files(files, openfn=open):
    """
    Generator which yield lines for each file from `files`.
    Input files can have format .ndjson, .json or .zip
    """
    for filename in files:
        try:
            mode = 'rb' if iszip(filename) else 'r'
            with openfn(filename, mode) as fd:
                if isndjson(filename):
                    for line in fd:
                        yield line.rstrip('\n')
                elif isjson(filename):
                    for line in prepare_json(fd.read()):
                        yield line
                elif iszip(filename):
                    with zipfile.ZipFile(fd) as zf:
                        for line in iter_lines_from_files(
                                zf.namelist(), make_zip_openfn(zf)):
                            yield line
                else:
                    logger.warning(
                        'Unknown file format for {0}'.format(filename))
        except FileNotFoundError:
            logger.warning('Cannot open file {0}'.format(filename))


def iter_lines_from_external(url):
    with requests.get(url, stream=True) as resp:
        if resp.status_code == 200:
            if isndjson(url):
                for line in resp.iter_lines():
                    yield line
            elif isjson(url):
                for line in prepare_json(resp.content):
                    yield line
            elif iszip(url):
                buf = io.BytesIO(resp.content)
                with zipfile.ZipFile(buf) as zf:
                    for line in iter_lines_from_files(
                            zf.namelist(), make_zip_openfn(zf)):
                        yield line
            else:
                logger.warning('Unknown file format for {0}'.format(url))

        else:
            logger.warning('Cannot fetch url {0}'.format(url))
