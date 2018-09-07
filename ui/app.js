
window.onload = function() {
  var mime = 'text/x-pgsql';
  // get mime type
  function runQuery(cm) {
    console.log(cm.getValue());
    return false;
  };
  window.editor = CodeMirror(document.getElementById('editor'), {
    mode: mime,
    theme: "zenburn",
    indentWithTabs: true,
    smartIndent: true,
    lineNumbers: true,
    matchBrackets : true,
    autofocus: true,
    extraKeys: {
      "Ctrl-Space": "autocomplete",
      "Ctrl-Enter": runQuery
    },
    hintOptions: {tables: {
      users: ["name", "score", "birthDate"],
      countries: ["name", "population", "size"]
    }}
  });
};

