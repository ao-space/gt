/*!
FullCalendar v5.11.5
Docs & License: https://fullcalendar.io/
(c) 2022 Adam Shaw
*/
'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var common = require('@fullcalendar/common');
var tslib = require('tslib');
;

/* An abstract class for the daygrid views, as well as month view. Renders one or more rows of day cells.
----------------------------------------------------------------------------------------------------------------------*/
// It is a manager for a Table subcomponent, which does most of the heavy lifting.
// It is responsible for managing width/height.
var TableView = /** @class */ (function (_super) {
    tslib.__extends(TableView, _super);
    function TableView() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.headerElRef = common.createRef();
        return _this;
    }
    TableView.prototype.renderSimpleLayout = function (headerRowContent, bodyContent) {
        var _a = this, props = _a.props, context = _a.context;
        var sections = [];
        var stickyHeaderDates = common.getStickyHeaderDates(context.options);
        if (headerRowContent) {
            sections.push({
                type: 'header',
                key: 'header',
                isSticky: stickyHeaderDates,
                chunk: {
                    elRef: this.headerElRef,
                    tableClassName: 'fc-col-header',
                    rowContent: headerRowContent,
                },
            });
        }
        sections.push({
            type: 'body',
            key: 'body',
            liquid: true,
            chunk: { content: bodyContent },
        });
        return (common.createElement(common.ViewRoot, { viewSpec: context.viewSpec }, function (rootElRef, classNames) { return (common.createElement("div", { ref: rootElRef, className: ['fc-daygrid'].concat(classNames).join(' ') },
            common.createElement(common.SimpleScrollGrid, { liquid: !props.isHeightAuto && !props.forPrint, collapsibleWidth: props.forPrint, cols: [] /* TODO: make optional? */, sections: sections }))); }));
    };
    TableView.prototype.renderHScrollLayout = function (headerRowContent, bodyContent, colCnt, dayMinWidth) {
        var ScrollGrid = this.context.pluginHooks.scrollGridImpl;
        if (!ScrollGrid) {
            throw new Error('No ScrollGrid implementation');
        }
        var _a = this, props = _a.props, context = _a.context;
        var stickyHeaderDates = !props.forPrint && common.getStickyHeaderDates(context.options);
        var stickyFooterScrollbar = !props.forPrint && common.getStickyFooterScrollbar(context.options);
        var sections = [];
        if (headerRowContent) {
            sections.push({
                type: 'header',
                key: 'header',
                isSticky: stickyHeaderDates,
                chunks: [{
                        key: 'main',
                        elRef: this.headerElRef,
                        tableClassName: 'fc-col-header',
                        rowContent: headerRowContent,
                    }],
            });
        }
        sections.push({
            type: 'body',
            key: 'body',
            liquid: true,
            chunks: [{
                    key: 'main',
                    content: bodyContent,
                }],
        });
        if (stickyFooterScrollbar) {
            sections.push({
                type: 'footer',
                key: 'footer',
                isSticky: true,
                chunks: [{
                        key: 'main',
                        content: common.renderScrollShim,
                    }],
            });
        }
        return (common.createElement(common.ViewRoot, { viewSpec: context.viewSpec }, function (rootElRef, classNames) { return (common.createElement("div", { ref: rootElRef, className: ['fc-daygrid'].concat(classNames).join(' ') },
            common.createElement(ScrollGrid, { liquid: !props.isHeightAuto && !props.forPrint, collapsibleWidth: props.forPrint, colGroups: [{ cols: [{ span: colCnt, minWidth: dayMinWidth }] }], sections: sections }))); }));
    };
    return TableView;
}(common.DateComponent));

function splitSegsByRow(segs, rowCnt) {
    var byRow = [];
    for (var i = 0; i < rowCnt; i += 1) {
        byRow[i] = [];
    }
    for (var _i = 0, segs_1 = segs; _i < segs_1.length; _i++) {
        var seg = segs_1[_i];
        byRow[seg.row].push(seg);
    }
    return byRow;
}
function splitSegsByFirstCol(segs, colCnt) {
    var byCol = [];
    for (var i = 0; i < colCnt; i += 1) {
        byCol[i] = [];
    }
    for (var _i = 0, segs_2 = segs; _i < segs_2.length; _i++) {
        var seg = segs_2[_i];
        byCol[seg.firstCol].push(seg);
    }
    return byCol;
}
function splitInteractionByRow(ui, rowCnt) {
    var byRow = [];
    if (!ui) {
        for (var i = 0; i < rowCnt; i += 1) {
            byRow[i] = null;
        }
    }
    else {
        for (var i = 0; i < rowCnt; i += 1) {
            byRow[i] = {
                affectedInstances: ui.affectedInstances,
                isEvent: ui.isEvent,
                segs: [],
            };
        }
        for (var _i = 0, _a = ui.segs; _i < _a.length; _i++) {
            var seg = _a[_i];
            byRow[seg.row].segs.push(seg);
        }
    }
    return byRow;
}

var TableCellTop = /** @class */ (function (_super) {
    tslib.__extends(TableCellTop, _super);
    function TableCellTop() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TableCellTop.prototype.render = function () {
        var props = this.props;
        var navLinkAttrs = common.buildNavLinkAttrs(this.context, props.date);
        return (common.createElement(common.DayCellContent, { date: props.date, dateProfile: props.dateProfile, todayRange: props.todayRange, showDayNumber: props.showDayNumber, extraHookProps: props.extraHookProps, defaultContent: renderTopInner }, function (innerElRef, innerContent) { return ((innerContent || props.forceDayTop) && (common.createElement("div", { className: "fc-daygrid-day-top", ref: innerElRef },
            common.createElement("a", tslib.__assign({ id: props.dayNumberId, className: "fc-daygrid-day-number" }, navLinkAttrs), innerContent || common.createElement(common.Fragment, null, "\u00A0"))))); }));
    };
    return TableCellTop;
}(common.BaseComponent));
function renderTopInner(props) {
    return props.dayNumberText;
}

var DEFAULT_TABLE_EVENT_TIME_FORMAT = common.createFormatter({
    hour: 'numeric',
    minute: '2-digit',
    omitZeroMinute: true,
    meridiem: 'narrow',
});
function hasListItemDisplay(seg) {
    var display = seg.eventRange.ui.display;
    return display === 'list-item' || (display === 'auto' &&
        !seg.eventRange.def.allDay &&
        seg.firstCol === seg.lastCol && // can't be multi-day
        seg.isStart && // "
        seg.isEnd // "
    );
}

var TableBlockEvent = /** @class */ (function (_super) {
    tslib.__extends(TableBlockEvent, _super);
    function TableBlockEvent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TableBlockEvent.prototype.render = function () {
        var props = this.props;
        return (common.createElement(common.StandardEvent, tslib.__assign({}, props, { extraClassNames: ['fc-daygrid-event', 'fc-daygrid-block-event', 'fc-h-event'], defaultTimeFormat: DEFAULT_TABLE_EVENT_TIME_FORMAT, defaultDisplayEventEnd: props.defaultDisplayEventEnd, disableResizing: !props.seg.eventRange.def.allDay })));
    };
    return TableBlockEvent;
}(common.BaseComponent));

var TableListItemEvent = /** @class */ (function (_super) {
    tslib.__extends(TableListItemEvent, _super);
    function TableListItemEvent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TableListItemEvent.prototype.render = function () {
        var _a = this, props = _a.props, context = _a.context;
        var timeFormat = context.options.eventTimeFormat || DEFAULT_TABLE_EVENT_TIME_FORMAT;
        var timeText = common.buildSegTimeText(props.seg, timeFormat, context, true, props.defaultDisplayEventEnd);
        return (common.createElement(common.EventRoot, { seg: props.seg, timeText: timeText, defaultContent: renderInnerContent, isDragging: props.isDragging, isResizing: false, isDateSelecting: false, isSelected: props.isSelected, isPast: props.isPast, isFuture: props.isFuture, isToday: props.isToday }, function (rootElRef, classNames, innerElRef, innerContent) { return ( // we don't use styles!
        common.createElement("a", tslib.__assign({ className: ['fc-daygrid-event', 'fc-daygrid-dot-event'].concat(classNames).join(' '), ref: rootElRef }, common.getSegAnchorAttrs(props.seg, context)), innerContent)); }));
    };
    return TableListItemEvent;
}(common.BaseComponent));
function renderInnerContent(innerProps) {
    return (common.createElement(common.Fragment, null,
        common.createElement("div", { className: "fc-daygrid-event-dot", style: { borderColor: innerProps.borderColor || innerProps.backgroundColor } }),
        innerProps.timeText && (common.createElement("div", { className: "fc-event-time" }, innerProps.timeText)),
        common.createElement("div", { className: "fc-event-title" }, innerProps.event.title || common.createElement(common.Fragment, null, "\u00A0"))));
}

var TableCellMoreLink = /** @class */ (function (_super) {
    tslib.__extends(TableCellMoreLink, _super);
    function TableCellMoreLink() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.compileSegs = common.memoize(compileSegs);
        return _this;
    }
    TableCellMoreLink.prototype.render = function () {
        var props = this.props;
        var _a = this.compileSegs(props.singlePlacements), allSegs = _a.allSegs, invisibleSegs = _a.invisibleSegs;
        return (common.createElement(common.MoreLinkRoot, { dateProfile: props.dateProfile, todayRange: props.todayRange, allDayDate: props.allDayDate, moreCnt: props.moreCnt, allSegs: allSegs, hiddenSegs: invisibleSegs, alignmentElRef: props.alignmentElRef, alignGridTop: props.alignGridTop, extraDateSpan: props.extraDateSpan, popoverContent: function () {
                var isForcedInvisible = (props.eventDrag ? props.eventDrag.affectedInstances : null) ||
                    (props.eventResize ? props.eventResize.affectedInstances : null) ||
                    {};
                return (common.createElement(common.Fragment, null, allSegs.map(function (seg) {
                    var instanceId = seg.eventRange.instance.instanceId;
                    return (common.createElement("div", { className: "fc-daygrid-event-harness", key: instanceId, style: {
                            visibility: isForcedInvisible[instanceId] ? 'hidden' : '',
                        } }, hasListItemDisplay(seg) ? (common.createElement(TableListItemEvent, tslib.__assign({ seg: seg, isDragging: false, isSelected: instanceId === props.eventSelection, defaultDisplayEventEnd: false }, common.getSegMeta(seg, props.todayRange)))) : (common.createElement(TableBlockEvent, tslib.__assign({ seg: seg, isDragging: false, isResizing: false, isDateSelecting: false, isSelected: instanceId === props.eventSelection, defaultDisplayEventEnd: false }, common.getSegMeta(seg, props.todayRange))))));
                })));
            } }, function (rootElRef, classNames, innerElRef, innerContent, handleClick, title, isExpanded, popoverId) { return (common.createElement("a", tslib.__assign({ ref: rootElRef, className: ['fc-daygrid-more-link'].concat(classNames).join(' '), title: title, "aria-expanded": isExpanded, "aria-controls": popoverId }, common.createAriaClickAttrs(handleClick)), innerContent)); }));
    };
    return TableCellMoreLink;
}(common.BaseComponent));
function compileSegs(singlePlacements) {
    var allSegs = [];
    var invisibleSegs = [];
    for (var _i = 0, singlePlacements_1 = singlePlacements; _i < singlePlacements_1.length; _i++) {
        var placement = singlePlacements_1[_i];
        allSegs.push(placement.seg);
        if (!placement.isVisible) {
            invisibleSegs.push(placement.seg);
        }
    }
    return { allSegs: allSegs, invisibleSegs: invisibleSegs };
}

var DEFAULT_WEEK_NUM_FORMAT = common.createFormatter({ week: 'narrow' });
var TableCell = /** @class */ (function (_super) {
    tslib.__extends(TableCell, _super);
    function TableCell() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.rootElRef = common.createRef();
        _this.state = {
            dayNumberId: common.getUniqueDomId(),
        };
        _this.handleRootEl = function (el) {
            common.setRef(_this.rootElRef, el);
            common.setRef(_this.props.elRef, el);
        };
        return _this;
    }
    TableCell.prototype.render = function () {
        var _a = this, context = _a.context, props = _a.props, state = _a.state, rootElRef = _a.rootElRef;
        var date = props.date, dateProfile = props.dateProfile;
        var navLinkAttrs = common.buildNavLinkAttrs(context, date, 'week');
        return (common.createElement(common.DayCellRoot, { date: date, dateProfile: dateProfile, todayRange: props.todayRange, showDayNumber: props.showDayNumber, extraHookProps: props.extraHookProps, elRef: this.handleRootEl }, function (dayElRef, dayClassNames, rootDataAttrs, isDisabled) { return (common.createElement("td", tslib.__assign({ ref: dayElRef, role: "gridcell", className: ['fc-daygrid-day'].concat(dayClassNames, props.extraClassNames || []).join(' ') }, rootDataAttrs, props.extraDataAttrs, (props.showDayNumber ? { 'aria-labelledby': state.dayNumberId } : {})),
            common.createElement("div", { className: "fc-daygrid-day-frame fc-scrollgrid-sync-inner", ref: props.innerElRef /* different from hook system! RENAME */ },
                props.showWeekNumber && (common.createElement(common.WeekNumberRoot, { date: date, defaultFormat: DEFAULT_WEEK_NUM_FORMAT }, function (weekElRef, weekClassNames, innerElRef, innerContent) { return (common.createElement("a", tslib.__assign({ ref: weekElRef, className: ['fc-daygrid-week-number'].concat(weekClassNames).join(' ') }, navLinkAttrs), innerContent)); })),
                !isDisabled && (common.createElement(TableCellTop, { date: date, dateProfile: dateProfile, showDayNumber: props.showDayNumber, dayNumberId: state.dayNumberId, forceDayTop: props.forceDayTop, todayRange: props.todayRange, extraHookProps: props.extraHookProps })),
                common.createElement("div", { className: "fc-daygrid-day-events", ref: props.fgContentElRef },
                    props.fgContent,
                    common.createElement("div", { className: "fc-daygrid-day-bottom", style: { marginTop: props.moreMarginTop } },
                        common.createElement(TableCellMoreLink, { allDayDate: date, singlePlacements: props.singlePlacements, moreCnt: props.moreCnt, alignmentElRef: rootElRef, alignGridTop: !props.showDayNumber, extraDateSpan: props.extraDateSpan, dateProfile: props.dateProfile, eventSelection: props.eventSelection, eventDrag: props.eventDrag, eventResize: props.eventResize, todayRange: props.todayRange }))),
                common.createElement("div", { className: "fc-daygrid-day-bg" }, props.bgContent)))); }));
    };
    return TableCell;
}(common.DateComponent));

function computeFgSegPlacement(segs, // assumed already sorted
dayMaxEvents, dayMaxEventRows, strictOrder, eventInstanceHeights, maxContentHeight, cells) {
    var hierarchy = new DayGridSegHierarchy();
    hierarchy.allowReslicing = true;
    hierarchy.strictOrder = strictOrder;
    if (dayMaxEvents === true || dayMaxEventRows === true) {
        hierarchy.maxCoord = maxContentHeight;
        hierarchy.hiddenConsumes = true;
    }
    else if (typeof dayMaxEvents === 'number') {
        hierarchy.maxStackCnt = dayMaxEvents;
    }
    else if (typeof dayMaxEventRows === 'number') {
        hierarchy.maxStackCnt = dayMaxEventRows;
        hierarchy.hiddenConsumes = true;
    }
    // create segInputs only for segs with known heights
    var segInputs = [];
    var unknownHeightSegs = [];
    for (var i = 0; i < segs.length; i += 1) {
        var seg = segs[i];
        var instanceId = seg.eventRange.instance.instanceId;
        var eventHeight = eventInstanceHeights[instanceId];
        if (eventHeight != null) {
            segInputs.push({
                index: i,
                thickness: eventHeight,
                span: {
                    start: seg.firstCol,
                    end: seg.lastCol + 1,
                },
            });
        }
        else {
            unknownHeightSegs.push(seg);
        }
    }
    var hiddenEntries = hierarchy.addSegs(segInputs);
    var segRects = hierarchy.toRects();
    var _a = placeRects(segRects, segs, cells), singleColPlacements = _a.singleColPlacements, multiColPlacements = _a.multiColPlacements, leftoverMargins = _a.leftoverMargins;
    var moreCnts = [];
    var moreMarginTops = [];
    // add segs with unknown heights
    for (var _i = 0, unknownHeightSegs_1 = unknownHeightSegs; _i < unknownHeightSegs_1.length; _i++) {
        var seg = unknownHeightSegs_1[_i];
        multiColPlacements[seg.firstCol].push({
            seg: seg,
            isVisible: false,
            isAbsolute: true,
            absoluteTop: 0,
            marginTop: 0,
        });
        for (var col = seg.firstCol; col <= seg.lastCol; col += 1) {
            singleColPlacements[col].push({
                seg: resliceSeg(seg, col, col + 1, cells),
                isVisible: false,
                isAbsolute: false,
                absoluteTop: 0,
                marginTop: 0,
            });
        }
    }
    // add the hidden entries
    for (var col = 0; col < cells.length; col += 1) {
        moreCnts.push(0);
    }
    for (var _b = 0, hiddenEntries_1 = hiddenEntries; _b < hiddenEntries_1.length; _b++) {
        var hiddenEntry = hiddenEntries_1[_b];
        var seg = segs[hiddenEntry.index];
        var hiddenSpan = hiddenEntry.span;
        multiColPlacements[hiddenSpan.start].push({
            seg: resliceSeg(seg, hiddenSpan.start, hiddenSpan.end, cells),
            isVisible: false,
            isAbsolute: true,
            absoluteTop: 0,
            marginTop: 0,
        });
        for (var col = hiddenSpan.start; col < hiddenSpan.end; col += 1) {
            moreCnts[col] += 1;
            singleColPlacements[col].push({
                seg: resliceSeg(seg, col, col + 1, cells),
                isVisible: false,
                isAbsolute: false,
                absoluteTop: 0,
                marginTop: 0,
            });
        }
    }
    // deal with leftover margins
    for (var col = 0; col < cells.length; col += 1) {
        moreMarginTops.push(leftoverMargins[col]);
    }
    return { singleColPlacements: singleColPlacements, multiColPlacements: multiColPlacements, moreCnts: moreCnts, moreMarginTops: moreMarginTops };
}
// rects ordered by top coord, then left
function placeRects(allRects, segs, cells) {
    var rectsByEachCol = groupRectsByEachCol(allRects, cells.length);
    var singleColPlacements = [];
    var multiColPlacements = [];
    var leftoverMargins = [];
    for (var col = 0; col < cells.length; col += 1) {
        var rects = rectsByEachCol[col];
        // compute all static segs in singlePlacements
        var singlePlacements = [];
        var currentHeight = 0;
        var currentMarginTop = 0;
        for (var _i = 0, rects_1 = rects; _i < rects_1.length; _i++) {
            var rect = rects_1[_i];
            var seg = segs[rect.index];
            singlePlacements.push({
                seg: resliceSeg(seg, col, col + 1, cells),
                isVisible: true,
                isAbsolute: false,
                absoluteTop: rect.levelCoord,
                marginTop: rect.levelCoord - currentHeight,
            });
            currentHeight = rect.levelCoord + rect.thickness;
        }
        // compute mixed static/absolute segs in multiPlacements
        var multiPlacements = [];
        currentHeight = 0;
        currentMarginTop = 0;
        for (var _a = 0, rects_2 = rects; _a < rects_2.length; _a++) {
            var rect = rects_2[_a];
            var seg = segs[rect.index];
            var isAbsolute = rect.span.end - rect.span.start > 1; // multi-column?
            var isFirstCol = rect.span.start === col;
            currentMarginTop += rect.levelCoord - currentHeight; // amount of space since bottom of previous seg
            currentHeight = rect.levelCoord + rect.thickness; // height will now be bottom of current seg
            if (isAbsolute) {
                currentMarginTop += rect.thickness;
                if (isFirstCol) {
                    multiPlacements.push({
                        seg: resliceSeg(seg, rect.span.start, rect.span.end, cells),
                        isVisible: true,
                        isAbsolute: true,
                        absoluteTop: rect.levelCoord,
                        marginTop: 0,
                    });
                }
            }
            else if (isFirstCol) {
                multiPlacements.push({
                    seg: resliceSeg(seg, rect.span.start, rect.span.end, cells),
                    isVisible: true,
                    isAbsolute: false,
                    absoluteTop: rect.levelCoord,
                    marginTop: currentMarginTop, // claim the margin
                });
                currentMarginTop = 0;
            }
        }
        singleColPlacements.push(singlePlacements);
        multiColPlacements.push(multiPlacements);
        leftoverMargins.push(currentMarginTop);
    }
    return { singleColPlacements: singleColPlacements, multiColPlacements: multiColPlacements, leftoverMargins: leftoverMargins };
}
function groupRectsByEachCol(rects, colCnt) {
    var rectsByEachCol = [];
    for (var col = 0; col < colCnt; col += 1) {
        rectsByEachCol.push([]);
    }
    for (var _i = 0, rects_3 = rects; _i < rects_3.length; _i++) {
        var rect = rects_3[_i];
        for (var col = rect.span.start; col < rect.span.end; col += 1) {
            rectsByEachCol[col].push(rect);
        }
    }
    return rectsByEachCol;
}
function resliceSeg(seg, spanStart, spanEnd, cells) {
    if (seg.firstCol === spanStart && seg.lastCol === spanEnd - 1) {
        return seg;
    }
    var eventRange = seg.eventRange;
    var origRange = eventRange.range;
    var slicedRange = common.intersectRanges(origRange, {
        start: cells[spanStart].date,
        end: common.addDays(cells[spanEnd - 1].date, 1),
    });
    return tslib.__assign(tslib.__assign({}, seg), { firstCol: spanStart, lastCol: spanEnd - 1, eventRange: {
            def: eventRange.def,
            ui: tslib.__assign(tslib.__assign({}, eventRange.ui), { durationEditable: false }),
            instance: eventRange.instance,
            range: slicedRange,
        }, isStart: seg.isStart && slicedRange.start.valueOf() === origRange.start.valueOf(), isEnd: seg.isEnd && slicedRange.end.valueOf() === origRange.end.valueOf() });
}
var DayGridSegHierarchy = /** @class */ (function (_super) {
    tslib.__extends(DayGridSegHierarchy, _super);
    function DayGridSegHierarchy() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        // config
        _this.hiddenConsumes = false;
        // allows us to keep hidden entries in the hierarchy so they take up space
        _this.forceHidden = {};
        return _this;
    }
    DayGridSegHierarchy.prototype.addSegs = function (segInputs) {
        var _this = this;
        var hiddenSegs = _super.prototype.addSegs.call(this, segInputs);
        var entriesByLevel = this.entriesByLevel;
        var excludeHidden = function (entry) { return !_this.forceHidden[common.buildEntryKey(entry)]; };
        // remove the forced-hidden segs
        for (var level = 0; level < entriesByLevel.length; level += 1) {
            entriesByLevel[level] = entriesByLevel[level].filter(excludeHidden);
        }
        return hiddenSegs;
    };
    DayGridSegHierarchy.prototype.handleInvalidInsertion = function (insertion, entry, hiddenEntries) {
        var _a = this, entriesByLevel = _a.entriesByLevel, forceHidden = _a.forceHidden;
        var touchingEntry = insertion.touchingEntry, touchingLevel = insertion.touchingLevel, touchingLateral = insertion.touchingLateral;
        if (this.hiddenConsumes && touchingEntry) {
            var touchingEntryId = common.buildEntryKey(touchingEntry);
            // if not already hidden
            if (!forceHidden[touchingEntryId]) {
                if (this.allowReslicing) {
                    var placeholderEntry = tslib.__assign(tslib.__assign({}, touchingEntry), { span: common.intersectSpans(touchingEntry.span, entry.span) });
                    var placeholderEntryId = common.buildEntryKey(placeholderEntry);
                    forceHidden[placeholderEntryId] = true;
                    entriesByLevel[touchingLevel][touchingLateral] = placeholderEntry; // replace touchingEntry with our placeholder
                    this.splitEntry(touchingEntry, entry, hiddenEntries); // split up the touchingEntry, reinsert it
                }
                else {
                    forceHidden[touchingEntryId] = true;
                    hiddenEntries.push(touchingEntry);
                }
            }
        }
        return _super.prototype.handleInvalidInsertion.call(this, insertion, entry, hiddenEntries);
    };
    return DayGridSegHierarchy;
}(common.SegHierarchy));

var TableRow = /** @class */ (function (_super) {
    tslib.__extends(TableRow, _super);
    function TableRow() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.cellElRefs = new common.RefMap(); // the <td>
        _this.frameElRefs = new common.RefMap(); // the fc-daygrid-day-frame
        _this.fgElRefs = new common.RefMap(); // the fc-daygrid-day-events
        _this.segHarnessRefs = new common.RefMap(); // indexed by "instanceId:firstCol"
        _this.rootElRef = common.createRef();
        _this.state = {
            framePositions: null,
            maxContentHeight: null,
            eventInstanceHeights: {},
        };
        return _this;
    }
    TableRow.prototype.render = function () {
        var _this = this;
        var _a = this, props = _a.props, state = _a.state, context = _a.context;
        var options = context.options;
        var colCnt = props.cells.length;
        var businessHoursByCol = splitSegsByFirstCol(props.businessHourSegs, colCnt);
        var bgEventSegsByCol = splitSegsByFirstCol(props.bgEventSegs, colCnt);
        var highlightSegsByCol = splitSegsByFirstCol(this.getHighlightSegs(), colCnt);
        var mirrorSegsByCol = splitSegsByFirstCol(this.getMirrorSegs(), colCnt);
        var _b = computeFgSegPlacement(common.sortEventSegs(props.fgEventSegs, options.eventOrder), props.dayMaxEvents, props.dayMaxEventRows, options.eventOrderStrict, state.eventInstanceHeights, state.maxContentHeight, props.cells), singleColPlacements = _b.singleColPlacements, multiColPlacements = _b.multiColPlacements, moreCnts = _b.moreCnts, moreMarginTops = _b.moreMarginTops;
        var isForcedInvisible = // TODO: messy way to compute this
         (props.eventDrag && props.eventDrag.affectedInstances) ||
            (props.eventResize && props.eventResize.affectedInstances) ||
            {};
        return (common.createElement("tr", { ref: this.rootElRef, role: "row" },
            props.renderIntro && props.renderIntro(),
            props.cells.map(function (cell, col) {
                var normalFgNodes = _this.renderFgSegs(col, props.forPrint ? singleColPlacements[col] : multiColPlacements[col], props.todayRange, isForcedInvisible);
                var mirrorFgNodes = _this.renderFgSegs(col, buildMirrorPlacements(mirrorSegsByCol[col], multiColPlacements), props.todayRange, {}, Boolean(props.eventDrag), Boolean(props.eventResize), false);
                return (common.createElement(TableCell, { key: cell.key, elRef: _this.cellElRefs.createRef(cell.key), innerElRef: _this.frameElRefs.createRef(cell.key) /* FF <td> problem, but okay to use for left/right. TODO: rename prop */, dateProfile: props.dateProfile, date: cell.date, showDayNumber: props.showDayNumbers, showWeekNumber: props.showWeekNumbers && col === 0, forceDayTop: props.showWeekNumbers /* even displaying weeknum for row, not necessarily day */, todayRange: props.todayRange, eventSelection: props.eventSelection, eventDrag: props.eventDrag, eventResize: props.eventResize, extraHookProps: cell.extraHookProps, extraDataAttrs: cell.extraDataAttrs, extraClassNames: cell.extraClassNames, extraDateSpan: cell.extraDateSpan, moreCnt: moreCnts[col], moreMarginTop: moreMarginTops[col], singlePlacements: singleColPlacements[col], fgContentElRef: _this.fgElRefs.createRef(cell.key), fgContent: ( // Fragment scopes the keys
                    common.createElement(common.Fragment, null,
                        common.createElement(common.Fragment, null, normalFgNodes),
                        common.createElement(common.Fragment, null, mirrorFgNodes))), bgContent: ( // Fragment scopes the keys
                    common.createElement(common.Fragment, null,
                        _this.renderFillSegs(highlightSegsByCol[col], 'highlight'),
                        _this.renderFillSegs(businessHoursByCol[col], 'non-business'),
                        _this.renderFillSegs(bgEventSegsByCol[col], 'bg-event'))) }));
            })));
    };
    TableRow.prototype.componentDidMount = function () {
        this.updateSizing(true);
    };
    TableRow.prototype.componentDidUpdate = function (prevProps, prevState) {
        var currentProps = this.props;
        this.updateSizing(!common.isPropsEqual(prevProps, currentProps));
    };
    TableRow.prototype.getHighlightSegs = function () {
        var props = this.props;
        if (props.eventDrag && props.eventDrag.segs.length) { // messy check
            return props.eventDrag.segs;
        }
        if (props.eventResize && props.eventResize.segs.length) { // messy check
            return props.eventResize.segs;
        }
        return props.dateSelectionSegs;
    };
    TableRow.prototype.getMirrorSegs = function () {
        var props = this.props;
        if (props.eventResize && props.eventResize.segs.length) { // messy check
            return props.eventResize.segs;
        }
        return [];
    };
    TableRow.prototype.renderFgSegs = function (col, segPlacements, todayRange, isForcedInvisible, isDragging, isResizing, isDateSelecting) {
        var context = this.context;
        var eventSelection = this.props.eventSelection;
        var framePositions = this.state.framePositions;
        var defaultDisplayEventEnd = this.props.cells.length === 1; // colCnt === 1
        var isMirror = isDragging || isResizing || isDateSelecting;
        var nodes = [];
        if (framePositions) {
            for (var _i = 0, segPlacements_1 = segPlacements; _i < segPlacements_1.length; _i++) {
                var placement = segPlacements_1[_i];
                var seg = placement.seg;
                var instanceId = seg.eventRange.instance.instanceId;
                var key = instanceId + ':' + col;
                var isVisible = placement.isVisible && !isForcedInvisible[instanceId];
                var isAbsolute = placement.isAbsolute;
                var left = '';
                var right = '';
                if (isAbsolute) {
                    if (context.isRtl) {
                        right = 0;
                        left = framePositions.lefts[seg.lastCol] - framePositions.lefts[seg.firstCol];
                    }
                    else {
                        left = 0;
                        right = framePositions.rights[seg.firstCol] - framePositions.rights[seg.lastCol];
                    }
                }
                /*
                known bug: events that are force to be list-item but span multiple days still take up space in later columns
                todo: in print view, for multi-day events, don't display title within non-start/end segs
                */
                nodes.push(common.createElement("div", { className: 'fc-daygrid-event-harness' + (isAbsolute ? ' fc-daygrid-event-harness-abs' : ''), key: key, ref: isMirror ? null : this.segHarnessRefs.createRef(key), style: {
                        visibility: isVisible ? '' : 'hidden',
                        marginTop: isAbsolute ? '' : placement.marginTop,
                        top: isAbsolute ? placement.absoluteTop : '',
                        left: left,
                        right: right,
                    } }, hasListItemDisplay(seg) ? (common.createElement(TableListItemEvent, tslib.__assign({ seg: seg, isDragging: isDragging, isSelected: instanceId === eventSelection, defaultDisplayEventEnd: defaultDisplayEventEnd }, common.getSegMeta(seg, todayRange)))) : (common.createElement(TableBlockEvent, tslib.__assign({ seg: seg, isDragging: isDragging, isResizing: isResizing, isDateSelecting: isDateSelecting, isSelected: instanceId === eventSelection, defaultDisplayEventEnd: defaultDisplayEventEnd }, common.getSegMeta(seg, todayRange))))));
            }
        }
        return nodes;
    };
    TableRow.prototype.renderFillSegs = function (segs, fillType) {
        var isRtl = this.context.isRtl;
        var todayRange = this.props.todayRange;
        var framePositions = this.state.framePositions;
        var nodes = [];
        if (framePositions) {
            for (var _i = 0, segs_1 = segs; _i < segs_1.length; _i++) {
                var seg = segs_1[_i];
                var leftRightCss = isRtl ? {
                    right: 0,
                    left: framePositions.lefts[seg.lastCol] - framePositions.lefts[seg.firstCol],
                } : {
                    left: 0,
                    right: framePositions.rights[seg.firstCol] - framePositions.rights[seg.lastCol],
                };
                nodes.push(common.createElement("div", { key: common.buildEventRangeKey(seg.eventRange), className: "fc-daygrid-bg-harness", style: leftRightCss }, fillType === 'bg-event' ?
                    common.createElement(common.BgEvent, tslib.__assign({ seg: seg }, common.getSegMeta(seg, todayRange))) :
                    common.renderFill(fillType)));
            }
        }
        return common.createElement.apply(void 0, tslib.__spreadArray([common.Fragment, {}], nodes));
    };
    TableRow.prototype.updateSizing = function (isExternalSizingChange) {
        var _a = this, props = _a.props, frameElRefs = _a.frameElRefs;
        if (!props.forPrint &&
            props.clientWidth !== null // positioning ready?
        ) {
            if (isExternalSizingChange) {
                var frameEls = props.cells.map(function (cell) { return frameElRefs.currentMap[cell.key]; });
                if (frameEls.length) {
                    var originEl = this.rootElRef.current;
                    this.setState({
                        framePositions: new common.PositionCache(originEl, frameEls, true, // isHorizontal
                        false),
                    });
                }
            }
            var oldInstanceHeights = this.state.eventInstanceHeights;
            var newInstanceHeights = this.queryEventInstanceHeights();
            var limitByContentHeight = props.dayMaxEvents === true || props.dayMaxEventRows === true;
            this.safeSetState({
                // HACK to prevent oscillations of events being shown/hidden from max-event-rows
                // Essentially, once you compute an element's height, never null-out.
                // TODO: always display all events, as visibility:hidden?
                eventInstanceHeights: tslib.__assign(tslib.__assign({}, oldInstanceHeights), newInstanceHeights),
                maxContentHeight: limitByContentHeight ? this.computeMaxContentHeight() : null,
            });
        }
    };
    TableRow.prototype.queryEventInstanceHeights = function () {
        var segElMap = this.segHarnessRefs.currentMap;
        var eventInstanceHeights = {};
        // get the max height amongst instance segs
        for (var key in segElMap) {
            var height = Math.round(segElMap[key].getBoundingClientRect().height);
            var instanceId = key.split(':')[0]; // deconstruct how renderFgSegs makes the key
            eventInstanceHeights[instanceId] = Math.max(eventInstanceHeights[instanceId] || 0, height);
        }
        return eventInstanceHeights;
    };
    TableRow.prototype.computeMaxContentHeight = function () {
        var firstKey = this.props.cells[0].key;
        var cellEl = this.cellElRefs.currentMap[firstKey];
        var fcContainerEl = this.fgElRefs.currentMap[firstKey];
        return cellEl.getBoundingClientRect().bottom - fcContainerEl.getBoundingClientRect().top;
    };
    TableRow.prototype.getCellEls = function () {
        var elMap = this.cellElRefs.currentMap;
        return this.props.cells.map(function (cell) { return elMap[cell.key]; });
    };
    return TableRow;
}(common.DateComponent));
TableRow.addStateEquality({
    eventInstanceHeights: common.isPropsEqual,
});
function buildMirrorPlacements(mirrorSegs, colPlacements) {
    if (!mirrorSegs.length) {
        return [];
    }
    var topsByInstanceId = buildAbsoluteTopHash(colPlacements); // TODO: cache this at first render?
    return mirrorSegs.map(function (seg) { return ({
        seg: seg,
        isVisible: true,
        isAbsolute: true,
        absoluteTop: topsByInstanceId[seg.eventRange.instance.instanceId],
        marginTop: 0,
    }); });
}
function buildAbsoluteTopHash(colPlacements) {
    var topsByInstanceId = {};
    for (var _i = 0, colPlacements_1 = colPlacements; _i < colPlacements_1.length; _i++) {
        var placements = colPlacements_1[_i];
        for (var _a = 0, placements_1 = placements; _a < placements_1.length; _a++) {
            var placement = placements_1[_a];
            topsByInstanceId[placement.seg.eventRange.instance.instanceId] = placement.absoluteTop;
        }
    }
    return topsByInstanceId;
}

var Table = /** @class */ (function (_super) {
    tslib.__extends(Table, _super);
    function Table() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.splitBusinessHourSegs = common.memoize(splitSegsByRow);
        _this.splitBgEventSegs = common.memoize(splitSegsByRow);
        _this.splitFgEventSegs = common.memoize(splitSegsByRow);
        _this.splitDateSelectionSegs = common.memoize(splitSegsByRow);
        _this.splitEventDrag = common.memoize(splitInteractionByRow);
        _this.splitEventResize = common.memoize(splitInteractionByRow);
        _this.rowRefs = new common.RefMap();
        _this.handleRootEl = function (rootEl) {
            _this.rootEl = rootEl;
            if (rootEl) {
                _this.context.registerInteractiveComponent(_this, {
                    el: rootEl,
                    isHitComboAllowed: _this.props.isHitComboAllowed,
                });
            }
            else {
                _this.context.unregisterInteractiveComponent(_this);
            }
        };
        return _this;
    }
    Table.prototype.render = function () {
        var _this = this;
        var props = this.props;
        var dateProfile = props.dateProfile, dayMaxEventRows = props.dayMaxEventRows, dayMaxEvents = props.dayMaxEvents, expandRows = props.expandRows;
        var rowCnt = props.cells.length;
        var businessHourSegsByRow = this.splitBusinessHourSegs(props.businessHourSegs, rowCnt);
        var bgEventSegsByRow = this.splitBgEventSegs(props.bgEventSegs, rowCnt);
        var fgEventSegsByRow = this.splitFgEventSegs(props.fgEventSegs, rowCnt);
        var dateSelectionSegsByRow = this.splitDateSelectionSegs(props.dateSelectionSegs, rowCnt);
        var eventDragByRow = this.splitEventDrag(props.eventDrag, rowCnt);
        var eventResizeByRow = this.splitEventResize(props.eventResize, rowCnt);
        var limitViaBalanced = dayMaxEvents === true || dayMaxEventRows === true;
        // if rows can't expand to fill fixed height, can't do balanced-height event limit
        // TODO: best place to normalize these options?
        if (limitViaBalanced && !expandRows) {
            limitViaBalanced = false;
            dayMaxEventRows = null;
            dayMaxEvents = null;
        }
        var classNames = [
            'fc-daygrid-body',
            limitViaBalanced ? 'fc-daygrid-body-balanced' : 'fc-daygrid-body-unbalanced',
            expandRows ? '' : 'fc-daygrid-body-natural', // will height of one row depend on the others?
        ];
        return (common.createElement("div", { className: classNames.join(' '), ref: this.handleRootEl, style: {
                // these props are important to give this wrapper correct dimensions for interactions
                // TODO: if we set it here, can we avoid giving to inner tables?
                width: props.clientWidth,
                minWidth: props.tableMinWidth,
            } },
            common.createElement(common.NowTimer, { unit: "day" }, function (nowDate, todayRange) { return (common.createElement(common.Fragment, null,
                common.createElement("table", { role: "presentation", className: "fc-scrollgrid-sync-table", style: {
                        width: props.clientWidth,
                        minWidth: props.tableMinWidth,
                        height: expandRows ? props.clientHeight : '',
                    } },
                    props.colGroupNode,
                    common.createElement("tbody", { role: "presentation" }, props.cells.map(function (cells, row) { return (common.createElement(TableRow, { ref: _this.rowRefs.createRef(row), key: cells.length
                            ? cells[0].date.toISOString() /* best? or put key on cell? or use diff formatter? */
                            : row // in case there are no cells (like when resource view is loading)
                        , showDayNumbers: rowCnt > 1, showWeekNumbers: props.showWeekNumbers, todayRange: todayRange, dateProfile: dateProfile, cells: cells, renderIntro: props.renderRowIntro, businessHourSegs: businessHourSegsByRow[row], eventSelection: props.eventSelection, bgEventSegs: bgEventSegsByRow[row].filter(isSegAllDay) /* hack */, fgEventSegs: fgEventSegsByRow[row], dateSelectionSegs: dateSelectionSegsByRow[row], eventDrag: eventDragByRow[row], eventResize: eventResizeByRow[row], dayMaxEvents: dayMaxEvents, dayMaxEventRows: dayMaxEventRows, clientWidth: props.clientWidth, clientHeight: props.clientHeight, forPrint: props.forPrint })); }))))); })));
    };
    // Hit System
    // ----------------------------------------------------------------------------------------------------
    Table.prototype.prepareHits = function () {
        this.rowPositions = new common.PositionCache(this.rootEl, this.rowRefs.collect().map(function (rowObj) { return rowObj.getCellEls()[0]; }), // first cell el in each row. TODO: not optimal
        false, true);
        this.colPositions = new common.PositionCache(this.rootEl, this.rowRefs.currentMap[0].getCellEls(), // cell els in first row
        true, // horizontal
        false);
    };
    Table.prototype.queryHit = function (positionLeft, positionTop) {
        var _a = this, colPositions = _a.colPositions, rowPositions = _a.rowPositions;
        var col = colPositions.leftToIndex(positionLeft);
        var row = rowPositions.topToIndex(positionTop);
        if (row != null && col != null) {
            var cell = this.props.cells[row][col];
            return {
                dateProfile: this.props.dateProfile,
                dateSpan: tslib.__assign({ range: this.getCellRange(row, col), allDay: true }, cell.extraDateSpan),
                dayEl: this.getCellEl(row, col),
                rect: {
                    left: colPositions.lefts[col],
                    right: colPositions.rights[col],
                    top: rowPositions.tops[row],
                    bottom: rowPositions.bottoms[row],
                },
                layer: 0,
            };
        }
        return null;
    };
    Table.prototype.getCellEl = function (row, col) {
        return this.rowRefs.currentMap[row].getCellEls()[col]; // TODO: not optimal
    };
    Table.prototype.getCellRange = function (row, col) {
        var start = this.props.cells[row][col].date;
        var end = common.addDays(start, 1);
        return { start: start, end: end };
    };
    return Table;
}(common.DateComponent));
function isSegAllDay(seg) {
    return seg.eventRange.def.allDay;
}

var DayTableSlicer = /** @class */ (function (_super) {
    tslib.__extends(DayTableSlicer, _super);
    function DayTableSlicer() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.forceDayIfListItem = true;
        return _this;
    }
    DayTableSlicer.prototype.sliceRange = function (dateRange, dayTableModel) {
        return dayTableModel.sliceRange(dateRange);
    };
    return DayTableSlicer;
}(common.Slicer));

var DayTable = /** @class */ (function (_super) {
    tslib.__extends(DayTable, _super);
    function DayTable() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.slicer = new DayTableSlicer();
        _this.tableRef = common.createRef();
        return _this;
    }
    DayTable.prototype.render = function () {
        var _a = this, props = _a.props, context = _a.context;
        return (common.createElement(Table, tslib.__assign({ ref: this.tableRef }, this.slicer.sliceProps(props, props.dateProfile, props.nextDayThreshold, context, props.dayTableModel), { dateProfile: props.dateProfile, cells: props.dayTableModel.cells, colGroupNode: props.colGroupNode, tableMinWidth: props.tableMinWidth, renderRowIntro: props.renderRowIntro, dayMaxEvents: props.dayMaxEvents, dayMaxEventRows: props.dayMaxEventRows, showWeekNumbers: props.showWeekNumbers, expandRows: props.expandRows, headerAlignElRef: props.headerAlignElRef, clientWidth: props.clientWidth, clientHeight: props.clientHeight, forPrint: props.forPrint })));
    };
    return DayTable;
}(common.DateComponent));

var DayTableView = /** @class */ (function (_super) {
    tslib.__extends(DayTableView, _super);
    function DayTableView() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.buildDayTableModel = common.memoize(buildDayTableModel);
        _this.headerRef = common.createRef();
        _this.tableRef = common.createRef();
        return _this;
    }
    DayTableView.prototype.render = function () {
        var _this = this;
        var _a = this.context, options = _a.options, dateProfileGenerator = _a.dateProfileGenerator;
        var props = this.props;
        var dayTableModel = this.buildDayTableModel(props.dateProfile, dateProfileGenerator);
        var headerContent = options.dayHeaders && (common.createElement(common.DayHeader, { ref: this.headerRef, dateProfile: props.dateProfile, dates: dayTableModel.headerDates, datesRepDistinctDays: dayTableModel.rowCnt === 1 }));
        var bodyContent = function (contentArg) { return (common.createElement(DayTable, { ref: _this.tableRef, dateProfile: props.dateProfile, dayTableModel: dayTableModel, businessHours: props.businessHours, dateSelection: props.dateSelection, eventStore: props.eventStore, eventUiBases: props.eventUiBases, eventSelection: props.eventSelection, eventDrag: props.eventDrag, eventResize: props.eventResize, nextDayThreshold: options.nextDayThreshold, colGroupNode: contentArg.tableColGroupNode, tableMinWidth: contentArg.tableMinWidth, dayMaxEvents: options.dayMaxEvents, dayMaxEventRows: options.dayMaxEventRows, showWeekNumbers: options.weekNumbers, expandRows: !props.isHeightAuto, headerAlignElRef: _this.headerElRef, clientWidth: contentArg.clientWidth, clientHeight: contentArg.clientHeight, forPrint: props.forPrint })); };
        return options.dayMinWidth
            ? this.renderHScrollLayout(headerContent, bodyContent, dayTableModel.colCnt, options.dayMinWidth)
            : this.renderSimpleLayout(headerContent, bodyContent);
    };
    return DayTableView;
}(TableView));
function buildDayTableModel(dateProfile, dateProfileGenerator) {
    var daySeries = new common.DaySeriesModel(dateProfile.renderRange, dateProfileGenerator);
    return new common.DayTableModel(daySeries, /year|month|week/.test(dateProfile.currentRangeUnit));
}

var TableDateProfileGenerator = /** @class */ (function (_super) {
    tslib.__extends(TableDateProfileGenerator, _super);
    function TableDateProfileGenerator() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    // Computes the date range that will be rendered.
    TableDateProfileGenerator.prototype.buildRenderRange = function (currentRange, currentRangeUnit, isRangeAllDay) {
        var dateEnv = this.props.dateEnv;
        var renderRange = _super.prototype.buildRenderRange.call(this, currentRange, currentRangeUnit, isRangeAllDay);
        var start = renderRange.start;
        var end = renderRange.end;
        var endOfWeek;
        // year and month views should be aligned with weeks. this is already done for week
        if (/^(year|month)$/.test(currentRangeUnit)) {
            start = dateEnv.startOfWeek(start);
            // make end-of-week if not already
            endOfWeek = dateEnv.startOfWeek(end);
            if (endOfWeek.valueOf() !== end.valueOf()) {
                end = common.addWeeks(endOfWeek, 1);
            }
        }
        // ensure 6 weeks
        if (this.props.monthMode &&
            this.props.fixedWeekCount) {
            var rowCnt = Math.ceil(// could be partial weeks due to hiddenDays
            common.diffWeeks(start, end));
            end = common.addWeeks(end, 6 - rowCnt);
        }
        return { start: start, end: end };
    };
    return TableDateProfileGenerator;
}(common.DateProfileGenerator));

var main = common.createPlugin({
    initialView: 'dayGridMonth',
    views: {
        dayGrid: {
            component: DayTableView,
            dateProfileGeneratorClass: TableDateProfileGenerator,
        },
        dayGridDay: {
            type: 'dayGrid',
            duration: { days: 1 },
        },
        dayGridWeek: {
            type: 'dayGrid',
            duration: { weeks: 1 },
        },
        dayGridMonth: {
            type: 'dayGrid',
            duration: { months: 1 },
            monthMode: true,
            fixedWeekCount: true,
        },
    },
});

exports.DayGridView = DayTableView;
exports.DayTable = DayTable;
exports.DayTableSlicer = DayTableSlicer;
exports.Table = Table;
exports.TableView = TableView;
exports.buildDayTableModel = buildDayTableModel;
exports.default = main;
