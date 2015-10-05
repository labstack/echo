(function() {
    var menu = document.querySelectorAll('.menu a');

    for (var i = 0; i < menu.length; i++) {
        var m = menu[i];
        if (location.href === m.href) {
            m.className += 'active';
        }
    }
})();
