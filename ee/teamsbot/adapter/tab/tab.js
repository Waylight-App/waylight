// The Teams tab (#334) shows nothing until the Teams library is initialised.
// The button opens the start page through Teams, so the desktop app hands it
// to the browser instead of opening it inside Teams.
(function () {
  var teams = window.microsoftTeams;
  if (!teams || !teams.app) {
    return; // opened outside Teams: the plain link works
  }
  teams.app.initialize().then(function () {
    var open = document.getElementById('open');
    if (!open || !teams.app.openLink) {
      return;
    }
    open.addEventListener('click', function (e) {
      e.preventDefault();
      teams.app.openLink(open.href).catch(function () {
        window.open(open.href, '_blank', 'noopener');
      });
    });
  });
})();
