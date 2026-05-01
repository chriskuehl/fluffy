var whitelistedKeys = [
    // arrow keys
    37, 38, 39, 40,

    // home/end
    35, 36,

    // page up/down
    33, 34,
];

function realSort(array) {
    // wtf javascript
    array.sort(function(a, b) { return a - b; });
}

function selectedLines() {
    /**
     * @return {array} list of currently selected line numbers
    */
    var h = window.location.hash || '#';
    if (h.substring(0, 1) === '#') {
        h = h.substring(1);
    }

    var maybe_selected = h.split(',');
    var selected = [];
    for (var i = 0; i < maybe_selected.length; i++) {
        var el = maybe_selected[i];
        if (el.substring(0, 1) !== 'L') {
            continue;
        }

        var parts = el.substring(1).split('-');
        if (parts.length == 1) {
            parts.push(parts[0]);
        }
        if (parts.length == 2) {
            parts[0] = parseInt(parts[0]);
            parts[1] = parseInt(parts[1]);

            for (var num = parts[0]; num <= parts[1]; num++) {
                if (num >= 1 && selected.indexOf(num) === -1) {
                    selected.push(num);
                }
            }
        }

    }
    realSort(selected);
    return selected;
}

function updateSelectedHash(selected) {  // XXX: mutates selected!
    function rangeFromNumber(num) {
        /*
         * Convert a string "num" to an Array with start and end bound.
         *
         * e.g.
         *   rangeFromNumber("42") => [42, 42]
         *   rangeFromNumber("42-48") => [42, 48]
         */
        var range = ('' + num).split('-').map(function(i) {
            return parseInt(i);
        });

        if (range.length == 1) {
            range.push(range[0]);
        }

        return range;
    }

    if (selected.length > 0) {
        realSort(selected);
        for (var i = selected.length - 1; i >= 0; i--) {
            var cur = rangeFromNumber(selected[i]);
            selected[i] = 'L' + selected[i];

            // see if we can combine with the previous one
            if (i > 0) {
                var prev = rangeFromNumber(selected[i - 1]);
                if (prev[1] == cur[0] - 1) {
                    selected[i - 1] = prev[0] + '-' + cur[1];
                    selected.splice(i, 1);
                }
            }
        }
        window.location.hash = '#' + selected.join(',');
    } else {
        // hack (for old browsers that don't support pushState) because
        // removing entirely causes a scroll to the top
        window.location.hash = '_';

        // for new browsers, use pushState to remove the hash entirely
        history.replaceState('', document.title, window.location.pathname);
    }
}

const lineNumbersFromClassList = (classList) => (
    [...classList]
        .map(c => c.replace('LL', 'line-'))
        .filter(c => c.startsWith('line-'))
        .map(c => parseInt(c.substring(5)))
);

function lineIdsFromAnchorClassList(classList) {
    return [...classList]
        .filter((c) => c.startsWith('LL'))
        .map((c) => parseInt(c.substring(2), 10))
        .filter((n) => !Number.isNaN(n));
}

var pasteWrapResizeObserver = null;
var pasteWrapSyncTimer = null;

function clearPasteWrapLineHeights() {
    $('.line-numbers > a').css('min-height', '');
    $('.text .highlight > pre > span').css('min-height', '');
}

function lineElementsByNum($container) {
    var lineElByNum = {};
    $container.find('.text .highlight > pre > span').each(function() {
        var lineNumbers = lineNumbersFromClassList(this.classList);
        for (var i = 0; i < lineNumbers.length; i++) {
            lineElByNum[lineNumbers[i]] = this;
        }
    });
    return lineElByNum;
}

function uniqueLineElements(lineElByNum, lineIds) {
    var seen = new Set();
    var elements = [];
    for (var i = 0; i < lineIds.length; i++) {
        var el = lineElByNum[lineIds[i]];
        if (el && !seen.has(el)) {
            seen.add(el);
            elements.push(el);
        }
    }
    return elements;
}

function totalLineElementHeight(elements) {
    var total = 0;
    for (var i = 0; i < elements.length; i++) {
        total += elements[i].getBoundingClientRect().height;
    }
    return total;
}

function syncPasteWrapLineHeightsForContainer($container) {
    var lineElByNum = lineElementsByNum($container);
    $container.find('.line-numbers > a').each(function() {
        var lineIds = lineIdsFromAnchorClassList(this.classList);
        var elements = uniqueLineElements(lineElByNum, lineIds);
        var total = totalLineElementHeight(elements);
        if (total > 0) {
            this.style.minHeight = total + 'px';
        } else {
            this.style.minHeight = '';
        }
    });
}

function syncPasteWrapLineHeightsForSideBySideDiff() {
    var containers = $('#paste .text-container').slice(0, 2);
    var paneRows = [];
    containers.each(function() {
        var $container = $(this);
        var lineElByNum = lineElementsByNum($container);
        var rows = [];
        $container.find('.line-numbers > a').each(function(index) {
            rows[index] = {
                anchor: this,
                elements: uniqueLineElements(
                    lineElByNum,
                    lineIdsFromAnchorClassList(this.classList),
                ),
            };
        });
        paneRows.push(rows);
    });

    var rowCount = Math.max(paneRows[0].length, paneRows[1].length);
    for (var rowIndex = 0; rowIndex < rowCount; rowIndex++) {
        var height = 0;
        for (var paneIndex = 0; paneIndex < paneRows.length; paneIndex++) {
            var row = paneRows[paneIndex][rowIndex];
            if (row) {
                height = Math.max(height, totalLineElementHeight(row.elements));
            }
        }

        var minHeight = height > 0 ? height + 'px' : '';
        for (var paneIndex = 0; paneIndex < paneRows.length; paneIndex++) {
            var row = paneRows[paneIndex][rowIndex];
            if (row) {
                row.anchor.style.minHeight = minHeight;
                for (var elIndex = 0; elIndex < row.elements.length; elIndex++) {
                    row.elements[elIndex].style.minHeight = minHeight;
                }
            }
        }
    }
}

function syncPasteWrapLineHeights() {
    if (!document.documentElement.classList.contains('paste-wrap-lines')) {
        clearPasteWrapLineHeights();
        return;
    }

    clearPasteWrapLineHeights();
    if (document.documentElement.classList.contains('diff-side-by-side')) {
        syncPasteWrapLineHeightsForSideBySideDiff();
        return;
    }

    $('.text-container').each(function() {
        syncPasteWrapLineHeightsForContainer($(this));
    });
}

function teardownPasteWrapResizeObserver() {
    if (pasteWrapResizeObserver !== null) {
        pasteWrapResizeObserver.disconnect();
        pasteWrapResizeObserver = null;
    }
}

function setupPasteWrapResizeObserver() {
    teardownPasteWrapResizeObserver();
    if (!document.documentElement.classList.contains('paste-wrap-lines')) {
        return;
    }
    if (typeof ResizeObserver === 'undefined') {
        return;
    }
    pasteWrapResizeObserver = new ResizeObserver(function() {
        schedulePasteWrapSync();
    });
    document.querySelectorAll('#paste .text-container .text').forEach(function(el) {
        pasteWrapResizeObserver.observe(el);
    });
}

function schedulePasteWrapSync() {
    if (pasteWrapSyncTimer !== null) {
        clearTimeout(pasteWrapSyncTimer);
    }
    pasteWrapSyncTimer = setTimeout(function() {
        pasteWrapSyncTimer = null;
        syncPasteWrapLineHeights();
    }, 50);
}

window.syncPasteWrapLineHeights = syncPasteWrapLineHeights;

$(document).ready(function() {
    var numbers = $('.line-numbers > a');
    var setState = -1;
    $(document).mouseup(function() {
        setState = -1;
    });

    function updateLineClasses(allSelectedLines, num, isSelected) {
        /**
         * Add or remove "selected" classes from the line number and its line.
         */
        var els = $('.LL' + num + ',.line-' + num);
        if (isSelected) {
            els.addClass('selected');
        } else {
            // Removing is a bit more complicated because a given line element
            // (either a line number element, or an actual line span in the
            // rendered text) may represent multiple lines after mapping. We
            // only want to remove the selected class if all lines for that
            // element are no longer selected.
            for (const el of els) {
                const selectedLines = lineNumbersFromClassList(el.classList)
                    .filter(line => allSelectedLines.includes(line));

                if (selectedLines.length === 0) {
                    el.classList.remove('selected');
                }
            }
        }
    }

    function maybeChangeState(el) {
        /**
         * Possibly change the state of a line number.
         *
         * Depending on the current mode of the drag (either highlight
         * or un-highlight), we update this row to match.
         */
        if (setState === -1) {
            return;
        }

        const selected = selectedLines();
        const lines = lineNumbersFromClassList(el[0].classList);

        for (const line of lines) {
            const idx = selected.indexOf(line);
            if (idx === -1 && setState === 1) {
                selected.push(line);
                updateSelectedHash([...selected]);
                updateLineClasses(selected, line, true);
            } else if (idx > -1 && setState === 0) {
                selected.splice(idx, 1);
                updateSelectedHash([...selected]);
                updateLineClasses(selected, line, false);
            }
        }
        return false;
    }

    numbers.on('mousedown', function(e) {
        setState = e.target.classList.contains("selected") ? 0 : 1;
        return maybeChangeState($(this));
    });
    numbers.on('click', function() { return false; });
    numbers.on('mouseenter', function(e) { return maybeChangeState($(this)); });
    numbers.on('dragstart', function(e) {
        return false;
    });

    $('#paste').on('keydown', function(e) {
        if (e.ctrlKey || e.metaKey) {
            if (e.which == 46 || e.which == 8) {  // backspace and delete
                return false;
            }
            return true;
        }
        if ($.inArray(e.which, whitelistedKeys) !== -1) {
            return true;
        }
        return false;
    }).on('cut paste', function(e) {
        return false;
    }).bind('dragover drop', function(e) {
        return false;
    });


    // highlight existing lines
    var selected = selectedLines();
    var hasMovedDown = false;
    for (var i = 0; i < selected.length; i++) {
        var line = selected[i];
        updateLineClasses(selected, line, true);

        if (!hasMovedDown) {
            var e = $('.LL' + line);
            hasMovedDown = true;
            if ($('body').scrollTop() == 0) {
                // Chrome needs body, Firefox needs html :\
                $('html, body').scrollTop(e.offset().top - 50);
            }
        }
    }

    $('#style').change(function() {
        changeStyleTo($(this).val());
        schedulePasteWrapSync();
    });

    if (document.documentElement.classList.contains('paste-wrap-lines')) {
        syncPasteWrapLineHeights();
        setupPasteWrapResizeObserver();
    }

    $(window).on('resize', schedulePasteWrapSync);

    $(window).on('load', function() {
        if (document.documentElement.classList.contains('paste-wrap-lines')) {
            syncPasteWrapLineHeights();
        }
    });

    var pasteWrapObserver = new MutationObserver(function() {
        if (document.documentElement.classList.contains('paste-wrap-lines')) {
            setupPasteWrapResizeObserver();
            schedulePasteWrapSync();
        } else {
            teardownPasteWrapResizeObserver();
            syncPasteWrapLineHeights();
        }
    });
    pasteWrapObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
    });
});
