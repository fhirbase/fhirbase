window.onload = function() {
  var mime = "text/x-pgsql";

  const escapeHtml = unsafe => {
    return unsafe
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  };

  const formatResultField = f => {
    if (typeof f === "object" && f !== null) {
      return "<pre>" + escapeHtml(JSON.stringify(f, null, 2)) + "</pre>";
    } else {
      return f;
    }
  };

  function runQuery(cm) {
    let q = cm.getValue();
    let url = new URL("/q", window.location);
    url.searchParams.append("query", q);

    fetch(url)
      .then(response => {
        return response.json();
      })
      .then(json => {
        console.log("Got results", json);
        let tbl =
          '<h3>Results</h3><table class="table table-striped table-bordered table-sm"><thead><tr>';

        json.columns.forEach(clmn => {
          tbl += "<th>" + clmn.Name + "</th>";
        });

        tbl += "</tr></thead><tbody>";

        json.rows.forEach(row => {
          tbl +=
            "<tr>" +
            row.map(f => "<td>" + formatResultField(f) + "</td>").join("") +
            "</tr>";
        });

        tbl += "</tbody></table>";

        console.log(tbl);

        document.getElementById("results").innerHTML = tbl;
      });
    return false;
  }

  window.submitQuery = () => {
    runQuery(window.editor);
  };

  window.editor = CodeMirror(document.getElementById("editor"), {
    mode: mime,
    theme: "duotone-light",
    indentWithTabs: true,
    smartIndent: true,
    lineNumbers: true,
    matchBrackets: true,
    value: "SELECT * FROM patient LIMIT 100;",
    autofocus: true,
    extraKeys: {
      "Ctrl-Space": "autocomplete",
      "Ctrl-Enter": runQuery
    },
    hintOptions: {
      tables: {
        users: ["name", "score", "birthDate"],
        countries: ["name", "population", "size"]
      }
    }
  });
};
