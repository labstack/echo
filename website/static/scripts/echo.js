(function() {
  // Add table class
  var tables = document.getElementsByTagName('table');
  for (var i = 0; i < tables.length; i++) {
    var t = tables[i];
    t.classList.add('w3-table-all');
  }

  // Add blockquote class
  var blockquotes = document.getElementsByTagName('blockquote');
  for (var i = 0; i < blockquotes.length; i++) {
    var bq = blockquotes[i];
    bq.classList.add('w3-panel');
    bq.classList.add('w3-leftbar');
    bq.classList.add('w3-light-grey');
  }

  // Add code class
  var codes = document.getElementsByTagName('code');
  for (var i = 0; i < codes.length; i++) {
    var c = codes[i];
    if (c.parentElement.nodeName !== 'PRE') {
      c.classList.add('w3-codespan');
    }
  }

  // Prism
  Prism.languages.sh = Prism.languages.bash;
})();

function openSidenav() {
  document.getElementById("sidenav").style.display = 'block';
}

function closeSidenav() {
  document.getElementById("sidenav").style.display = 'none';
}
