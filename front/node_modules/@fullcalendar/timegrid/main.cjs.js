/*!
FullCalendar v5.11.5
Docs & License: https://fullcalendar.io/
(c) 2022 Adam Shaw
*/
'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var common = require('@fullcalendar/common');
var tslib = require('tslib');
var daygrid = require('@fullcalendar/daygrid');
;

var AllDaySplitter = /** @class */ (function (_super) {
    tslib.__extends(AllDaySplitter, _super);
    function AllDaySplitter() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    AllDaySplitter.prototype.getKeyInfo = function () {
        return {
            allDay: {},
            timed: {},
        };
    };
    AllDaySplitter.prototype.getKeysForDateSpan = function (dateSpan) {
        if (dateSpan.allDay) {
            return ['allDay'];
        }
        return ['timed'];
    };
    AllDaySplitter.prototype.getKeysForEventDef = function (eventDef) {
        if (!eventDef.allDay) {
            return ['timed'];
        }
        if (common.hasBgRendering(eventDef)) {
            return ['timed', 'allDay'];
        }
        return ['allDay'];
    };
    return AllDaySplitter;
}(common.Splitter));

var DEFAULT_SLAT_LABEL_FORMAT = common.createFormatter({
    hour: 'numeric',
    minute: '2-digit',
    omitZeroMinute: true,
    meridiem: 'short',
});
function TimeColsAxisCell(props) {
    var classNames = [
        'fc-timegrid-slot',
        'fc-timegrid-slot-label',
        props.isLabeled ? 'fc-scrollgrid-shrink' : 'fc-timegrid-slot-minor',
    ];
    return (common.createElement(common.ViewContextType.Consumer, null, function (context) {
        if (!props.isLabeled) {
            return (common.createElement("td", { className: classNames.join(' '), "data-time": props.isoTimeStr }));
        }
        var dateEnv = context.dateEnv, options = context.options, viewApi = context.viewApi;
        var labelFormat = // TODO: fully pre-parse
         options.slotLabelFormat == null ? DEFAULT_SLAT_LABEL_FORMAT :
            Array.isArray(options.slotLabelFormat) ? common.createFormatter(options.slotLabelFormat[0]) :
                common.createFormatter(options.slotLabelFormat);
        var hookProps = {
            level: 0,
            time: props.time,
            date: dateEnv.toDate(props.date),
            view: viewApi,
            text: dateEnv.format(props.date, labelFormat),
        };
        return (common.createElement(common.RenderHook, { hookProps: hookProps, classNames: options.slotLabelClassNames, content: options.slotLabelContent, defaultContent: renderInnerContent, didMount: options.slotLabelDidMount, willUnmount: options.slotLabelWillUnmount }, function (rootElRef, customClassNames, innerElRef, innerContent) { return (common.createElement("td", { ref: rootElRef, className: classNames.concat(customClassNames).join(' '), "data-time": props.isoTimeStr },
            common.createElement("div", { className: "fc-timegrid-slot-label-frame fc-scrollgrid-shrink-frame" },
                common.createElement("div", { className: "fc-timegrid-slot-label-cushion fc-scrollgrid-shrink-cushion", ref: innerElRef }, innerContent)))); }));
    }));
}
function renderInnerContent(props) {
    return props.text;
}

var TimeBodyAxis = /** @class */ (function (_super) {
    tslib.__extends(TimeBodyAxis, _super);
    function TimeBodyAxis() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TimeBodyAxis.prototype.render = function () {
        return this.props.slatMetas.map(function (slatMeta) { return (common.createElement("tr", { key: slatMeta.key },
            common.createElement(TimeColsAxisCell, tslib.__assign({}, slatMeta)))); });
    };
    return TimeBodyAxis;
}(common.BaseComponent));

var DEFAULT_WEEK_NUM_FORMAT = common.createFormatter({ week: 'short' });
var AUTO_ALL_DAY_MAX_EVENT_ROWS = 5;
var TimeColsView = /** @class */ (function (_super) {
    tslib.__extends(TimeColsView, _super);
    function TimeColsView() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.allDaySplitter = new AllDaySplitter(); // for use by subclasses
        _this.headerElRef = common.createRef();
        _this.rootElRef = common.createRef();
        _this.scrollerElRef = common.createRef();
        _this.state = {
            slatCoords: null,
        };
        _this.handleScrollTopRequest = function (scrollTop) {
            var scrollerEl = _this.scrollerElRef.current;
            if (scrollerEl) { // TODO: not sure how this could ever be null. weirdness with the reducer
                scrollerEl.scrollTop = scrollTop;
            }
        };
        /* Header Render Methods
        ------------------------------------------------------------------------------------------------------------------*/
        _this.renderHeadAxis = function (rowKey, frameHeight) {
            if (frameHeight === void 0) { frameHeight = ''; }
            var options = _this.context.options;
            var dateProfile = _this.props.dateProfile;
            var range = dateProfile.renderRange;
            var dayCnt = common.diffDays(range.start, range.end);
            var navLinkAttrs = (dayCnt === 1) // only do in day views (to avoid doing in week views that dont need it)
                ? common.buildNavLinkAttrs(_this.context, range.start, 'week')
                : {};
            if (options.weekNumbers && rowKey === 'day') {
                return (common.createElement(common.WeekNumberRoot, { date: range.start, defaultFormat: DEFAULT_WEEK_NUM_FORMAT }, function (rootElRef, classNames, innerElRef, innerContent) { return (common.createElement("th", { ref: rootElRef, "aria-hidden": true, className: [
                        'fc-timegrid-axis',
                        'fc-scrollgrid-shrink',
                    ].concat(classNames).join(' ') },
                    common.createElement("div", { className: "fc-timegrid-axis-frame fc-scrollgrid-shrink-frame fc-timegrid-axis-frame-liquid", style: { height: frameHeight } },
                        common.createElement("a", tslib.__assign({ ref: innerElRef, className: "fc-timegrid-axis-cushion fc-scrollgrid-shrink-cushion fc-scrollgrid-sync-inner" }, navLinkAttrs), innerContent)))); }));
            }
            return (common.createElement("th", { "aria-hidden": true, className: "fc-timegrid-axis" },
                common.createElement("div", { className: "fc-timegrid-axis-frame", style: { height: frameHeight } })));
        };
        /* Table Component Render Methods
        ------------------------------------------------------------------------------------------------------------------*/
        // only a one-way height sync. we don't send the axis inner-content height to the DayGrid,
        // but DayGrid still needs to have classNames on inner elements in order to measure.
        _this.renderTableRowAxis = function (rowHeight) {
            var _a = _this.context, options = _a.options, viewApi = _a.viewApi;
            var hookProps = {
                text: options.allDayText,
                view: viewApi,
            };
            return (
            // TODO: make reusable hook. used in list view too
            common.createElement(common.RenderHook, { hookProps: hookProps, classNames: options.allDayClassNames, content: options.allDayContent, defaultContent: renderAllDayInner, didMount: options.allDayDidMount, willUnmount: options.allDayWillUnmount }, function (rootElRef, classNames, innerElRef, innerContent) { return (common.createElement("td", { ref: rootElRef, "aria-hidden": true, className: [
                    'fc-timegrid-axis',
                    'fc-scrollgrid-shrink',
                ].concat(classNames).join(' ') },
                common.createElement("div", { className: 'fc-timegrid-axis-frame fc-scrollgrid-shrink-frame' + (rowHeight == null ? ' fc-timegrid-axis-frame-liquid' : ''), style: { height: rowHeight } },
                    common.createElement("span", { className: "fc-timegrid-axis-cushion fc-scrollgrid-shrink-cushion fc-scrollgrid-sync-inner", ref: innerElRef }, innerContent)))); }));
        };
        _this.handleSlatCoords = function (slatCoords) {
            _this.setState({ slatCoords: slatCoords });
        };
        return _this;
    }
    // rendering
    // ----------------------------------------------------------------------------------------------------
    TimeColsView.prototype.renderSimpleLayout = function (headerRowContent, allDayContent, timeContent) {
        var _a = this, context = _a.context, props = _a.props;
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
        if (allDayContent) {
            sections.push({
                type: 'body',
                key: 'all-day',
                chunk: { content: allDayContent },
            });
            sections.push({
                type: 'body',
                key: 'all-day-divider',
                outerContent: ( // TODO: rename to cellContent so don't need to define <tr>?
                common.createElement("tr", { role: "presentation", className: "fc-scrollgrid-section" },
                    common.createElement("td", { className: 'fc-timegrid-divider ' + context.theme.getClass('tableCellShaded') }))),
            });
        }
        sections.push({
            type: 'body',
            key: 'body',
            liquid: true,
            expandRows: Boolean(context.options.expandRows),
            chunk: {
                scrollerElRef: this.scrollerElRef,
                content: timeContent,
            },
        });
        return (common.createElement(common.ViewRoot, { viewSpec: context.viewSpec, elRef: this.rootElRef }, function (rootElRef, classNames) { return (common.createElement("div", { className: ['fc-timegrid'].concat(classNames).join(' '), ref: rootElRef },
            common.createElement(common.SimpleScrollGrid, { liquid: !props.isHeightAuto && !props.forPrint, collapsibleWidth: props.forPrint, cols: [{ width: 'shrink' }], sections: sections }))); }));
    };
    TimeColsView.prototype.renderHScrollLayout = function (headerRowContent, allDayContent, timeContent, colCnt, dayMinWidth, slatMetas, slatCoords) {
        var _this = this;
        var ScrollGrid = this.context.pluginHooks.scrollGridImpl;
        if (!ScrollGrid) {
            throw new Error('No ScrollGrid implementation');
        }
        var _a = this, context = _a.context, props = _a.props;
        var stickyHeaderDates = !props.forPrint && common.getStickyHeaderDates(context.options);
        var stickyFooterScrollbar = !props.forPrint && common.getStickyFooterScrollbar(context.options);
        var sections = [];
        if (headerRowContent) {
            sections.push({
                type: 'header',
                key: 'header',
                isSticky: stickyHeaderDates,
                syncRowHeights: true,
                chunks: [
                    {
                        key: 'axis',
                        rowContent: function (arg) { return (common.createElement("tr", { role: "presentation" }, _this.renderHeadAxis('day', arg.rowSyncHeights[0]))); },
                    },
                    {
                        key: 'cols',
                        elRef: this.headerElRef,
                        tableClassName: 'fc-col-header',
                        rowContent: headerRowContent,
                    },
                ],
            });
        }
        if (allDayContent) {
            sections.push({
                type: 'body',
                key: 'all-day',
                syncRowHeights: true,
                chunks: [
                    {
                        key: 'axis',
                        rowContent: function (contentArg) { return (common.createElement("tr", { role: "presentation" }, _this.renderTableRowAxis(contentArg.rowSyncHeights[0]))); },
                    },
                    {
                        key: 'cols',
                        content: allDayContent,
                    },
                ],
            });
            sections.push({
                key: 'all-day-divider',
                type: 'body',
                outerContent: ( // TODO: rename to cellContent so don't need to define <tr>?
                common.createElement("tr", { role: "presentation", className: "fc-scrollgrid-section" },
                    common.createElement("td", { colSpan: 2, className: 'fc-timegrid-divider ' + context.theme.getClass('tableCellShaded') }))),
            });
        }
        var isNowIndicator = context.options.nowIndicator;
        sections.push({
            type: 'body',
            key: 'body',
            liquid: true,
            expandRows: Boolean(context.options.expandRows),
            chunks: [
                {
                    key: 'axis',
                    content: function (arg) { return (
                    // TODO: make this now-indicator arrow more DRY with TimeColsContent
                    common.createElement("div", { className: "fc-timegrid-axis-chunk" },
                        common.createElement("table", { "aria-hidden": true, style: { height: arg.expandRows ? arg.clientHeight : '' } },
                            arg.tableColGroupNode,
                            common.createElement("tbody", null,
                                common.createElement(TimeBodyAxis, { slatMetas: slatMetas }))),
                        common.createElement("div", { className: "fc-timegrid-now-indicator-container" },
                            common.createElement(common.NowTimer, { unit: isNowIndicator ? 'minute' : 'day' /* hacky */ }, function (nowDate) {
                                var nowIndicatorTop = isNowIndicator &&
                                    slatCoords &&
                                    slatCoords.safeComputeTop(nowDate); // might return void
                                if (typeof nowIndicatorTop === 'number') {
                                    return (common.createElement(common.NowIndicatorRoot, { isAxis: true, date: nowDate }, function (rootElRef, classNames, innerElRef, innerContent) { return (common.createElement("div", { ref: rootElRef, className: ['fc-timegrid-now-indicator-arrow'].concat(classNames).join(' '), style: { top: nowIndicatorTop } }, innerContent)); }));
                                }
                                return null;
                            })))); },
                },
                {
                    key: 'cols',
                    scrollerElRef: this.scrollerElRef,
                    content: timeContent,
                },
            ],
        });
        if (stickyFooterScrollbar) {
            sections.push({
                key: 'footer',
                type: 'footer',
                isSticky: true,
                chunks: [
                    {
                        key: 'axis',
                        content: common.renderScrollShim,
                    },
                    {
                        key: 'cols',
                        content: common.renderScrollShim,
                    },
                ],
            });
        }
        return (common.createElement(common.ViewRoot, { viewSpec: context.viewSpec, elRef: this.rootElRef }, function (rootElRef, classNames) { return (common.createElement("div", { className: ['fc-timegrid'].concat(classNames).join(' '), ref: rootElRef },
            common.createElement(ScrollGrid, { liquid: !props.isHeightAuto && !props.forPrint, collapsibleWidth: false, colGroups: [
                    { width: 'shrink', cols: [{ width: 'shrink' }] },
                    { cols: [{ span: colCnt, minWidth: dayMinWidth }] },
                ], sections: sections }))); }));
    };
    /* Dimensions
    ------------------------------------------------------------------------------------------------------------------*/
    TimeColsView.prototype.getAllDayMaxEventProps = function () {
        var _a = this.context.options, dayMaxEvents = _a.dayMaxEvents, dayMaxEventRows = _a.dayMaxEventRows;
        if (dayMaxEvents === true || dayMaxEventRows === true) { // is auto?
            dayMaxEvents = undefined;
            dayMaxEventRows = AUTO_ALL_DAY_MAX_EVENT_ROWS; // make sure "auto" goes to a real number
        }
        return { dayMaxEvents: dayMaxEvents, dayMaxEventRows: dayMaxEventRows };
    };
    return TimeColsView;
}(common.DateComponent));
function renderAllDayInner(hookProps) {
    return hookProps.text;
}

var TimeColsSlatsCoords = /** @class */ (function () {
    function TimeColsSlatsCoords(positions, dateProfile, slotDuration) {
        this.positions = positions;
        this.dateProfile = dateProfile;
        this.slotDuration = slotDuration;
    }
    TimeColsSlatsCoords.prototype.safeComputeTop = function (date) {
        var dateProfile = this.dateProfile;
        if (common.rangeContainsMarker(dateProfile.currentRange, date)) {
            var startOfDayDate = common.startOfDay(date);
            var timeMs = date.valueOf() - startOfDayDate.valueOf();
            if (timeMs >= common.asRoughMs(dateProfile.slotMinTime) &&
                timeMs < common.asRoughMs(dateProfile.slotMaxTime)) {
                return this.computeTimeTop(common.createDuration(timeMs));
            }
        }
        return null;
    };
    // Computes the top coordinate, relative to the bounds of the grid, of the given date.
    // A `startOfDayDate` must be given for avoiding ambiguity over how to treat midnight.
    TimeColsSlatsCoords.prototype.computeDateTop = function (when, startOfDayDate) {
        if (!startOfDayDate) {
            startOfDayDate = common.startOfDay(when);
        }
        return this.computeTimeTop(common.createDuration(when.valueOf() - startOfDayDate.valueOf()));
    };
    // Computes the top coordinate, relative to the bounds of the grid, of the given time (a Duration).
    // This is a makeshify way to compute the time-top. Assumes all slatMetas dates are uniform.
    // Eventually allow computation with arbirary slat dates.
    TimeColsSlatsCoords.prototype.computeTimeTop = function (duration) {
        var _a = this, positions = _a.positions, dateProfile = _a.dateProfile;
        var len = positions.els.length;
        // floating-point value of # of slots covered
        var slatCoverage = (duration.milliseconds - common.asRoughMs(dateProfile.slotMinTime)) / common.asRoughMs(this.slotDuration);
        var slatIndex;
        var slatRemainder;
        // compute a floating-point number for how many slats should be progressed through.
        // from 0 to number of slats (inclusive)
        // constrained because slotMinTime/slotMaxTime might be customized.
        slatCoverage = Math.max(0, slatCoverage);
        slatCoverage = Math.min(len, slatCoverage);
        // an integer index of the furthest whole slat
        // from 0 to number slats (*exclusive*, so len-1)
        slatIndex = Math.floor(slatCoverage);
        slatIndex = Math.min(slatIndex, len - 1);
        // how much further through the slatIndex slat (from 0.0-1.0) must be covered in addition.
        // could be 1.0 if slatCoverage is covering *all* the slots
        slatRemainder = slatCoverage - slatIndex;
        return positions.tops[slatIndex] +
            positions.getHeight(slatIndex) * slatRemainder;
    };
    return TimeColsSlatsCoords;
}());

var TimeColsSlatsBody = /** @class */ (function (_super) {
    tslib.__extends(TimeColsSlatsBody, _super);
    function TimeColsSlatsBody() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TimeColsSlatsBody.prototype.render = function () {
        var _a = this, props = _a.props, context = _a.context;
        var options = context.options;
        var slatElRefs = props.slatElRefs;
        return (common.createElement("tbody", null, props.slatMetas.map(function (slatMeta, i) {
            var hookProps = {
                time: slatMeta.time,
                date: context.dateEnv.toDate(slatMeta.date),
                view: context.viewApi,
            };
            var classNames = [
                'fc-timegrid-slot',
                'fc-timegrid-slot-lane',
                slatMeta.isLabeled ? '' : 'fc-timegrid-slot-minor',
            ];
            return (common.createElement("tr", { key: slatMeta.key, ref: slatElRefs.createRef(slatMeta.key) },
                props.axis && (common.createElement(TimeColsAxisCell, tslib.__assign({}, slatMeta))),
                common.createElement(common.RenderHook, { hookProps: hookProps, classNames: options.slotLaneClassNames, content: options.slotLaneContent, didMount: options.slotLaneDidMount, willUnmount: options.slotLaneWillUnmount }, function (rootElRef, customClassNames, innerElRef, innerContent) { return (common.createElement("td", { ref: rootElRef, className: classNames.concat(customClassNames).join(' '), "data-time": slatMeta.isoTimeStr }, innerContent)); })));
        })));
    };
    return TimeColsSlatsBody;
}(common.BaseComponent));

/*
for the horizontal "slats" that run width-wise. Has a time axis on a side. Depends on RTL.
*/
var TimeColsSlats = /** @class */ (function (_super) {
    tslib.__extends(TimeColsSlats, _super);
    function TimeColsSlats() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.rootElRef = common.createRef();
        _this.slatElRefs = new common.RefMap();
        return _this;
    }
    TimeColsSlats.prototype.render = function () {
        var _a = this, props = _a.props, context = _a.context;
        return (common.createElement("div", { ref: this.rootElRef, className: "fc-timegrid-slots" },
            common.createElement("table", { "aria-hidden": true, className: context.theme.getClass('table'), style: {
                    minWidth: props.tableMinWidth,
                    width: props.clientWidth,
                    height: props.minHeight,
                } },
                props.tableColGroupNode /* relies on there only being a single <col> for the axis */,
                common.createElement(TimeColsSlatsBody, { slatElRefs: this.slatElRefs, axis: props.axis, slatMetas: props.slatMetas }))));
    };
    TimeColsSlats.prototype.componentDidMount = function () {
        this.updateSizing();
    };
    TimeColsSlats.prototype.componentDidUpdate = function () {
        this.updateSizing();
    };
    TimeColsSlats.prototype.componentWillUnmount = function () {
        if (this.props.onCoords) {
            this.props.onCoords(null);
        }
    };
    TimeColsSlats.prototype.updateSizing = function () {
        var _a = this, context = _a.context, props = _a.props;
        if (props.onCoords &&
            props.clientWidth !== null // means sizing has stabilized
        ) {
            var rootEl = this.rootElRef.current;
            if (rootEl.offsetHeight) { // not hidden by css
                props.onCoords(new TimeColsSlatsCoords(new common.PositionCache(this.rootElRef.current, collectSlatEls(this.slatElRefs.currentMap, props.slatMetas), false, true), this.props.dateProfile, context.options.slotDuration));
            }
        }
    };
    return TimeColsSlats;
}(common.BaseComponent));
function collectSlatEls(elMap, slatMetas) {
    return slatMetas.map(function (slatMeta) { return elMap[slatMeta.key]; });
}

function splitSegsByCol(segs, colCnt) {
    var segsByCol = [];
    var i;
    for (i = 0; i < colCnt; i += 1) {
        segsByCol.push([]);
    }
    if (segs) {
        for (i = 0; i < segs.length; i += 1) {
            segsByCol[segs[i].col].push(segs[i]);
        }
    }
    return segsByCol;
}
function splitInteractionByCol(ui, colCnt) {
    var byRow = [];
    if (!ui) {
        for (var i = 0; i < colCnt; i += 1) {
            byRow[i] = null;
        }
    }
    else {
        for (var i = 0; i < colCnt; i += 1) {
            byRow[i] = {
                affectedInstances: ui.affectedInstances,
                isEvent: ui.isEvent,
                segs: [],
            };
        }
        for (var _i = 0, _a = ui.segs; _i < _a.length; _i++) {
            var seg = _a[_i];
            byRow[seg.col].segs.push(seg);
        }
    }
    return byRow;
}

var TimeColMoreLink = /** @class */ (function (_super) {
    tslib.__extends(TimeColMoreLink, _super);
    function TimeColMoreLink() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.rootElRef = common.createRef();
        return _this;
    }
    TimeColMoreLink.prototype.render = function () {
        var _this = this;
        var props = this.props;
        return (common.createElement(common.MoreLinkRoot, { allDayDate: null, moreCnt: props.hiddenSegs.length, allSegs: props.hiddenSegs, hiddenSegs: props.hiddenSegs, alignmentElRef: this.rootElRef, defaultContent: renderMoreLinkInner, extraDateSpan: props.extraDateSpan, dateProfile: props.dateProfile, todayRange: props.todayRange, popoverContent: function () { return renderPlainFgSegs(props.hiddenSegs, props); } }, function (rootElRef, classNames, innerElRef, innerContent, handleClick, title, isExpanded, popoverId) { return (common.createElement("a", { ref: function (el) {
                common.setRef(rootElRef, el);
                common.setRef(_this.rootElRef, el);
            }, className: ['fc-timegrid-more-link'].concat(classNames).join(' '), style: { top: props.top, bottom: props.bottom }, onClick: handleClick, title: title, "aria-expanded": isExpanded, "aria-controls": popoverId },
            common.createElement("div", { ref: innerElRef, className: "fc-timegrid-more-link-inner fc-sticky" }, innerContent))); }));
    };
    return TimeColMoreLink;
}(common.BaseComponent));
function renderMoreLinkInner(props) {
    return props.shortText;
}

// segInputs assumed sorted
function buildPositioning(segInputs, strictOrder, maxStackCnt) {
    var hierarchy = new common.SegHierarchy();
    if (strictOrder != null) {
        hierarchy.strictOrder = strictOrder;
    }
    if (maxStackCnt != null) {
        hierarchy.maxStackCnt = maxStackCnt;
    }
    var hiddenEntries = hierarchy.addSegs(segInputs);
    var hiddenGroups = common.groupIntersectingEntries(hiddenEntries);
    var web = buildWeb(hierarchy);
    web = stretchWeb(web, 1); // all levelCoords/thickness will have 0.0-1.0
    var segRects = webToRects(web);
    return { segRects: segRects, hiddenGroups: hiddenGroups };
}
function buildWeb(hierarchy) {
    var entriesByLevel = hierarchy.entriesByLevel;
    var buildNode = cacheable(function (level, lateral) { return level + ':' + lateral; }, function (level, lateral) {
        var siblingRange = findNextLevelSegs(hierarchy, level, lateral);
        var nextLevelRes = buildNodes(siblingRange, buildNode);
        var entry = entriesByLevel[level][lateral];
        return [
            tslib.__assign(tslib.__assign({}, entry), { nextLevelNodes: nextLevelRes[0] }),
            entry.thickness + nextLevelRes[1], // the pressure builds
        ];
    });
    return buildNodes(entriesByLevel.length
        ? { level: 0, lateralStart: 0, lateralEnd: entriesByLevel[0].length }
        : null, buildNode)[0];
}
function buildNodes(siblingRange, buildNode) {
    if (!siblingRange) {
        return [[], 0];
    }
    var level = siblingRange.level, lateralStart = siblingRange.lateralStart, lateralEnd = siblingRange.lateralEnd;
    var lateral = lateralStart;
    var pairs = [];
    while (lateral < lateralEnd) {
        pairs.push(buildNode(level, lateral));
        lateral += 1;
    }
    pairs.sort(cmpDescPressures);
    return [
        pairs.map(extractNode),
        pairs[0][1], // first item's pressure
    ];
}
function cmpDescPressures(a, b) {
    return b[1] - a[1];
}
function extractNode(a) {
    return a[0];
}
function findNextLevelSegs(hierarchy, subjectLevel, subjectLateral) {
    var levelCoords = hierarchy.levelCoords, entriesByLevel = hierarchy.entriesByLevel;
    var subjectEntry = entriesByLevel[subjectLevel][subjectLateral];
    var afterSubject = levelCoords[subjectLevel] + subjectEntry.thickness;
    var levelCnt = levelCoords.length;
    var level = subjectLevel;
    // skip past levels that are too high up
    for (; level < levelCnt && levelCoords[level] < afterSubject; level += 1)
        ; // do nothing
    for (; level < levelCnt; level += 1) {
        var entries = entriesByLevel[level];
        var entry = void 0;
        var searchIndex = common.binarySearch(entries, subjectEntry.span.start, common.getEntrySpanEnd);
        var lateralStart = searchIndex[0] + searchIndex[1]; // if exact match (which doesn't collide), go to next one
        var lateralEnd = lateralStart;
        while ( // loop through entries that horizontally intersect
        (entry = entries[lateralEnd]) && // but not past the whole seg list
            entry.span.start < subjectEntry.span.end) {
            lateralEnd += 1;
        }
        if (lateralStart < lateralEnd) {
            return { level: level, lateralStart: lateralStart, lateralEnd: lateralEnd };
        }
    }
    return null;
}
function stretchWeb(topLevelNodes, totalThickness) {
    var stretchNode = cacheable(function (node, startCoord, prevThickness) { return common.buildEntryKey(node); }, function (node, startCoord, prevThickness) {
        var nextLevelNodes = node.nextLevelNodes, thickness = node.thickness;
        var allThickness = thickness + prevThickness;
        var thicknessFraction = thickness / allThickness;
        var endCoord;
        var newChildren = [];
        if (!nextLevelNodes.length) {
            endCoord = totalThickness;
        }
        else {
            for (var _i = 0, nextLevelNodes_1 = nextLevelNodes; _i < nextLevelNodes_1.length; _i++) {
                var childNode = nextLevelNodes_1[_i];
                if (endCoord === undefined) {
                    var res = stretchNode(childNode, startCoord, allThickness);
                    endCoord = res[0];
                    newChildren.push(res[1]);
                }
                else {
                    var res = stretchNode(childNode, endCoord, 0);
                    newChildren.push(res[1]);
                }
            }
        }
        var newThickness = (endCoord - startCoord) * thicknessFraction;
        return [endCoord - newThickness, tslib.__assign(tslib.__assign({}, node), { thickness: newThickness, nextLevelNodes: newChildren })];
    });
    return topLevelNodes.map(function (node) { return stretchNode(node, 0, 0)[1]; });
}
// not sorted in any particular order
function webToRects(topLevelNodes) {
    var rects = [];
    var processNode = cacheable(function (node, levelCoord, stackDepth) { return common.buildEntryKey(node); }, function (node, levelCoord, stackDepth) {
        var rect = tslib.__assign(tslib.__assign({}, node), { levelCoord: levelCoord,
            stackDepth: stackDepth, stackForward: 0 });
        rects.push(rect);
        return (rect.stackForward = processNodes(node.nextLevelNodes, levelCoord + node.thickness, stackDepth + 1) + 1);
    });
    function processNodes(nodes, levelCoord, stackDepth) {
        var stackForward = 0;
        for (var _i = 0, nodes_1 = nodes; _i < nodes_1.length; _i++) {
            var node = nodes_1[_i];
            stackForward = Math.max(processNode(node, levelCoord, stackDepth), stackForward);
        }
        return stackForward;
    }
    processNodes(topLevelNodes, 0, 0);
    return rects; // TODO: sort rects by levelCoord to be consistent with toRects?
}
// TODO: move to general util
function cacheable(keyFunc, workFunc) {
    var cache = {};
    return function () {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        var key = keyFunc.apply(void 0, args);
        return (key in cache)
            ? cache[key]
            : (cache[key] = workFunc.apply(void 0, args));
    };
}

function computeSegVCoords(segs, colDate, slatCoords, eventMinHeight) {
    if (slatCoords === void 0) { slatCoords = null; }
    if (eventMinHeight === void 0) { eventMinHeight = 0; }
    var vcoords = [];
    if (slatCoords) {
        for (var i = 0; i < segs.length; i += 1) {
            var seg = segs[i];
            var spanStart = slatCoords.computeDateTop(seg.start, colDate);
            var spanEnd = Math.max(spanStart + (eventMinHeight || 0), // :(
            slatCoords.computeDateTop(seg.end, colDate));
            vcoords.push({
                start: Math.round(spanStart),
                end: Math.round(spanEnd), //
            });
        }
    }
    return vcoords;
}
function computeFgSegPlacements(segs, segVCoords, // might not have for every seg
eventOrderStrict, eventMaxStack) {
    var segInputs = [];
    var dumbSegs = []; // segs without coords
    for (var i = 0; i < segs.length; i += 1) {
        var vcoords = segVCoords[i];
        if (vcoords) {
            segInputs.push({
                index: i,
                thickness: 1,
                span: vcoords,
            });
        }
        else {
            dumbSegs.push(segs[i]);
        }
    }
    var _a = buildPositioning(segInputs, eventOrderStrict, eventMaxStack), segRects = _a.segRects, hiddenGroups = _a.hiddenGroups;
    var segPlacements = [];
    for (var _i = 0, segRects_1 = segRects; _i < segRects_1.length; _i++) {
        var segRect = segRects_1[_i];
        segPlacements.push({
            seg: segs[segRect.index],
            rect: segRect,
        });
    }
    for (var _b = 0, dumbSegs_1 = dumbSegs; _b < dumbSegs_1.length; _b++) {
        var dumbSeg = dumbSegs_1[_b];
        segPlacements.push({ seg: dumbSeg, rect: null });
    }
    return { segPlacements: segPlacements, hiddenGroups: hiddenGroups };
}

var DEFAULT_TIME_FORMAT = common.createFormatter({
    hour: 'numeric',
    minute: '2-digit',
    meridiem: false,
});
var TimeColEvent = /** @class */ (function (_super) {
    tslib.__extends(TimeColEvent, _super);
    function TimeColEvent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TimeColEvent.prototype.render = function () {
        var classNames = [
            'fc-timegrid-event',
            'fc-v-event',
        ];
        if (this.props.isShort) {
            classNames.push('fc-timegrid-event-short');
        }
        return (common.createElement(common.StandardEvent, tslib.__assign({}, this.props, { defaultTimeFormat: DEFAULT_TIME_FORMAT, extraClassNames: classNames })));
    };
    return TimeColEvent;
}(common.BaseComponent));

var TimeColMisc = /** @class */ (function (_super) {
    tslib.__extends(TimeColMisc, _super);
    function TimeColMisc() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    TimeColMisc.prototype.render = function () {
        var props = this.props;
        return (common.createElement(common.DayCellContent, { date: props.date, dateProfile: props.dateProfile, todayRange: props.todayRange, extraHookProps: props.extraHookProps }, function (innerElRef, innerContent) { return (innerContent &&
            common.createElement("div", { className: "fc-timegrid-col-misc", ref: innerElRef }, innerContent)); }));
    };
    return TimeColMisc;
}(common.BaseComponent));

var TimeCol = /** @class */ (function (_super) {
    tslib.__extends(TimeCol, _super);
    function TimeCol() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.sortEventSegs = common.memoize(common.sortEventSegs);
        return _this;
    }
    // TODO: memoize event-placement?
    TimeCol.prototype.render = function () {
        var _this = this;
        var _a = this, props = _a.props, context = _a.context;
        var isSelectMirror = context.options.selectMirror;
        var mirrorSegs = (props.eventDrag && props.eventDrag.segs) ||
            (props.eventResize && props.eventResize.segs) ||
            (isSelectMirror && props.dateSelectionSegs) ||
            [];
        var interactionAffectedInstances = // TODO: messy way to compute this
         (props.eventDrag && props.eventDrag.affectedInstances) ||
            (props.eventResize && props.eventResize.affectedInstances) ||
            {};
        var sortedFgSegs = this.sortEventSegs(props.fgEventSegs, context.options.eventOrder);
        return (common.createElement(common.DayCellRoot, { elRef: props.elRef, date: props.date, dateProfile: props.dateProfile, todayRange: props.todayRange, extraHookProps: props.extraHookProps }, function (rootElRef, classNames, dataAttrs) { return (common.createElement("td", tslib.__assign({ ref: rootElRef, role: "gridcell", className: ['fc-timegrid-col'].concat(classNames, props.extraClassNames || []).join(' ') }, dataAttrs, props.extraDataAttrs),
            common.createElement("div", { className: "fc-timegrid-col-frame" },
                common.createElement("div", { className: "fc-timegrid-col-bg" },
                    _this.renderFillSegs(props.businessHourSegs, 'non-business'),
                    _this.renderFillSegs(props.bgEventSegs, 'bg-event'),
                    _this.renderFillSegs(props.dateSelectionSegs, 'highlight')),
                common.createElement("div", { className: "fc-timegrid-col-events" }, _this.renderFgSegs(sortedFgSegs, interactionAffectedInstances, false, false, false)),
                common.createElement("div", { className: "fc-timegrid-col-events" }, _this.renderFgSegs(mirrorSegs, {}, Boolean(props.eventDrag), Boolean(props.eventResize), Boolean(isSelectMirror))),
                common.createElement("div", { className: "fc-timegrid-now-indicator-container" }, _this.renderNowIndicator(props.nowIndicatorSegs)),
                common.createElement(TimeColMisc, { date: props.date, dateProfile: props.dateProfile, todayRange: props.todayRange, extraHookProps: props.extraHookProps })))); }));
    };
    TimeCol.prototype.renderFgSegs = function (sortedFgSegs, segIsInvisible, isDragging, isResizing, isDateSelecting) {
        var props = this.props;
        if (props.forPrint) {
            return renderPlainFgSegs(sortedFgSegs, props);
        }
        return this.renderPositionedFgSegs(sortedFgSegs, segIsInvisible, isDragging, isResizing, isDateSelecting);
    };
    TimeCol.prototype.renderPositionedFgSegs = function (segs, // if not mirror, needs to be sorted
    segIsInvisible, isDragging, isResizing, isDateSelecting) {
        var _this = this;
        var _a = this.context.options, eventMaxStack = _a.eventMaxStack, eventShortHeight = _a.eventShortHeight, eventOrderStrict = _a.eventOrderStrict, eventMinHeight = _a.eventMinHeight;
        var _b = this.props, date = _b.date, slatCoords = _b.slatCoords, eventSelection = _b.eventSelection, todayRange = _b.todayRange, nowDate = _b.nowDate;
        var isMirror = isDragging || isResizing || isDateSelecting;
        var segVCoords = computeSegVCoords(segs, date, slatCoords, eventMinHeight);
        var _c = computeFgSegPlacements(segs, segVCoords, eventOrderStrict, eventMaxStack), segPlacements = _c.segPlacements, hiddenGroups = _c.hiddenGroups;
        return (common.createElement(common.Fragment, null,
            this.renderHiddenGroups(hiddenGroups, segs),
            segPlacements.map(function (segPlacement) {
                var seg = segPlacement.seg, rect = segPlacement.rect;
                var instanceId = seg.eventRange.instance.instanceId;
                var isVisible = isMirror || Boolean(!segIsInvisible[instanceId] && rect);
                var vStyle = computeSegVStyle(rect && rect.span);
                var hStyle = (!isMirror && rect) ? _this.computeSegHStyle(rect) : { left: 0, right: 0 };
                var isInset = Boolean(rect) && rect.stackForward > 0;
                var isShort = Boolean(rect) && (rect.span.end - rect.span.start) < eventShortHeight; // look at other places for this problem
                return (common.createElement("div", { className: 'fc-timegrid-event-harness' +
                        (isInset ? ' fc-timegrid-event-harness-inset' : ''), key: instanceId, style: tslib.__assign(tslib.__assign({ visibility: isVisible ? '' : 'hidden' }, vStyle), hStyle) },
                    common.createElement(TimeColEvent, tslib.__assign({ seg: seg, isDragging: isDragging, isResizing: isResizing, isDateSelecting: isDateSelecting, isSelected: instanceId === eventSelection, isShort: isShort }, common.getSegMeta(seg, todayRange, nowDate)))));
            })));
    };
    // will already have eventMinHeight applied because segInputs already had it
    TimeCol.prototype.renderHiddenGroups = function (hiddenGroups, segs) {
        var _a = this.props, extraDateSpan = _a.extraDateSpan, dateProfile = _a.dateProfile, todayRange = _a.todayRange, nowDate = _a.nowDate, eventSelection = _a.eventSelection, eventDrag = _a.eventDrag, eventResize = _a.eventResize;
        return (common.createElement(common.Fragment, null, hiddenGroups.map(function (hiddenGroup) {
            var positionCss = computeSegVStyle(hiddenGroup.span);
            var hiddenSegs = compileSegsFromEntries(hiddenGroup.entries, segs);
            return (common.createElement(TimeColMoreLink, { key: common.buildIsoString(common.computeEarliestSegStart(hiddenSegs)), hiddenSegs: hiddenSegs, top: positionCss.top, bottom: positionCss.bottom, extraDateSpan: extraDateSpan, dateProfile: dateProfile, todayRange: todayRange, nowDate: nowDate, eventSelection: eventSelection, eventDrag: eventDrag, eventResize: eventResize }));
        })));
    };
    TimeCol.prototype.renderFillSegs = function (segs, fillType) {
        var _a = this, props = _a.props, context = _a.context;
        var segVCoords = computeSegVCoords(segs, props.date, props.slatCoords, context.options.eventMinHeight); // don't assume all populated
        var children = segVCoords.map(function (vcoords, i) {
            var seg = segs[i];
            return (common.createElement("div", { key: common.buildEventRangeKey(seg.eventRange), className: "fc-timegrid-bg-harness", style: computeSegVStyle(vcoords) }, fillType === 'bg-event' ?
                common.createElement(common.BgEvent, tslib.__assign({ seg: seg }, common.getSegMeta(seg, props.todayRange, props.nowDate))) :
                common.renderFill(fillType)));
        });
        return common.createElement(common.Fragment, null, children);
    };
    TimeCol.prototype.renderNowIndicator = function (segs) {
        var _a = this.props, slatCoords = _a.slatCoords, date = _a.date;
        if (!slatCoords) {
            return null;
        }
        return segs.map(function (seg, i) { return (common.createElement(common.NowIndicatorRoot, { isAxis: false, date: date, 
            // key doesn't matter. will only ever be one
            key: i }, function (rootElRef, classNames, innerElRef, innerContent) { return (common.createElement("div", { ref: rootElRef, className: ['fc-timegrid-now-indicator-line'].concat(classNames).join(' '), style: { top: slatCoords.computeDateTop(seg.start, date) } }, innerContent)); })); });
    };
    TimeCol.prototype.computeSegHStyle = function (segHCoords) {
        var _a = this.context, isRtl = _a.isRtl, options = _a.options;
        var shouldOverlap = options.slotEventOverlap;
        var nearCoord = segHCoords.levelCoord; // the left side if LTR. the right side if RTL. floating-point
        var farCoord = segHCoords.levelCoord + segHCoords.thickness; // the right side if LTR. the left side if RTL. floating-point
        var left; // amount of space from left edge, a fraction of the total width
        var right; // amount of space from right edge, a fraction of the total width
        if (shouldOverlap) {
            // double the width, but don't go beyond the maximum forward coordinate (1.0)
            farCoord = Math.min(1, nearCoord + (farCoord - nearCoord) * 2);
        }
        if (isRtl) {
            left = 1 - farCoord;
            right = nearCoord;
        }
        else {
            left = nearCoord;
            right = 1 - farCoord;
        }
        var props = {
            zIndex: segHCoords.stackDepth + 1,
            left: left * 100 + '%',
            right: right * 100 + '%',
        };
        if (shouldOverlap && !segHCoords.stackForward) {
            // add padding to the edge so that forward stacked events don't cover the resizer's icon
            props[isRtl ? 'marginLeft' : 'marginRight'] = 10 * 2; // 10 is a guesstimate of the icon's width
        }
        return props;
    };
    return TimeCol;
}(common.BaseComponent));
function renderPlainFgSegs(sortedFgSegs, _a) {
    var todayRange = _a.todayRange, nowDate = _a.nowDate, eventSelection = _a.eventSelection, eventDrag = _a.eventDrag, eventResize = _a.eventResize;
    var hiddenInstances = (eventDrag ? eventDrag.affectedInstances : null) ||
        (eventResize ? eventResize.affectedInstances : null) ||
        {};
    return (common.createElement(common.Fragment, null, sortedFgSegs.map(function (seg) {
        var instanceId = seg.eventRange.instance.instanceId;
        return (common.createElement("div", { key: instanceId, style: { visibility: hiddenInstances[instanceId] ? 'hidden' : '' } },
            common.createElement(TimeColEvent, tslib.__assign({ seg: seg, isDragging: false, isResizing: false, isDateSelecting: false, isSelected: instanceId === eventSelection, isShort: false }, common.getSegMeta(seg, todayRange, nowDate)))));
    })));
}
function computeSegVStyle(segVCoords) {
    if (!segVCoords) {
        return { top: '', bottom: '' };
    }
    return {
        top: segVCoords.start,
        bottom: -segVCoords.end,
    };
}
function compileSegsFromEntries(segEntries, allSegs) {
    return segEntries.map(function (segEntry) { return allSegs[segEntry.index]; });
}

var TimeColsContent = /** @class */ (function (_super) {
    tslib.__extends(TimeColsContent, _super);
    function TimeColsContent() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.splitFgEventSegs = common.memoize(splitSegsByCol);
        _this.splitBgEventSegs = common.memoize(splitSegsByCol);
        _this.splitBusinessHourSegs = common.memoize(splitSegsByCol);
        _this.splitNowIndicatorSegs = common.memoize(splitSegsByCol);
        _this.splitDateSelectionSegs = common.memoize(splitSegsByCol);
        _this.splitEventDrag = common.memoize(splitInteractionByCol);
        _this.splitEventResize = common.memoize(splitInteractionByCol);
        _this.rootElRef = common.createRef();
        _this.cellElRefs = new common.RefMap();
        return _this;
    }
    TimeColsContent.prototype.render = function () {
        var _this = this;
        var _a = this, props = _a.props, context = _a.context;
        var nowIndicatorTop = context.options.nowIndicator &&
            props.slatCoords &&
            props.slatCoords.safeComputeTop(props.nowDate); // might return void
        var colCnt = props.cells.length;
        var fgEventSegsByRow = this.splitFgEventSegs(props.fgEventSegs, colCnt);
        var bgEventSegsByRow = this.splitBgEventSegs(props.bgEventSegs, colCnt);
        var businessHourSegsByRow = this.splitBusinessHourSegs(props.businessHourSegs, colCnt);
        var nowIndicatorSegsByRow = this.splitNowIndicatorSegs(props.nowIndicatorSegs, colCnt);
        var dateSelectionSegsByRow = this.splitDateSelectionSegs(props.dateSelectionSegs, colCnt);
        var eventDragByRow = this.splitEventDrag(props.eventDrag, colCnt);
        var eventResizeByRow = this.splitEventResize(props.eventResize, colCnt);
        return (common.createElement("div", { className: "fc-timegrid-cols", ref: this.rootElRef },
            common.createElement("table", { role: "presentation", style: {
                    minWidth: props.tableMinWidth,
                    width: props.clientWidth,
                } },
                props.tableColGroupNode,
                common.createElement("tbody", { role: "presentation" },
                    common.createElement("tr", { role: "row" },
                        props.axis && (common.createElement("td", { "aria-hidden": true, className: "fc-timegrid-col fc-timegrid-axis" },
                            common.createElement("div", { className: "fc-timegrid-col-frame" },
                                common.createElement("div", { className: "fc-timegrid-now-indicator-container" }, typeof nowIndicatorTop === 'number' && (common.createElement(common.NowIndicatorRoot, { isAxis: true, date: props.nowDate }, function (rootElRef, classNames, innerElRef, innerContent) { return (common.createElement("div", { ref: rootElRef, className: ['fc-timegrid-now-indicator-arrow'].concat(classNames).join(' '), style: { top: nowIndicatorTop } }, innerContent)); })))))),
                        props.cells.map(function (cell, i) { return (common.createElement(TimeCol, { key: cell.key, elRef: _this.cellElRefs.createRef(cell.key), dateProfile: props.dateProfile, date: cell.date, nowDate: props.nowDate, todayRange: props.todayRange, extraHookProps: cell.extraHookProps, extraDataAttrs: cell.extraDataAttrs, extraClassNames: cell.extraClassNames, extraDateSpan: cell.extraDateSpan, fgEventSegs: fgEventSegsByRow[i], bgEventSegs: bgEventSegsByRow[i], businessHourSegs: businessHourSegsByRow[i], nowIndicatorSegs: nowIndicatorSegsByRow[i], dateSelectionSegs: dateSelectionSegsByRow[i], eventDrag: eventDragByRow[i], eventResize: eventResizeByRow[i], slatCoords: props.slatCoords, eventSelection: props.eventSelection, forPrint: props.forPrint })); }))))));
    };
    TimeColsContent.prototype.componentDidMount = function () {
        this.updateCoords();
    };
    TimeColsContent.prototype.componentDidUpdate = function () {
        this.updateCoords();
    };
    TimeColsContent.prototype.updateCoords = function () {
        var props = this.props;
        if (props.onColCoords &&
            props.clientWidth !== null // means sizing has stabilized
        ) {
            props.onColCoords(new common.PositionCache(this.rootElRef.current, collectCellEls(this.cellElRefs.currentMap, props.cells), true, // horizontal
            false));
        }
    };
    return TimeColsContent;
}(common.BaseComponent));
function collectCellEls(elMap, cells) {
    return cells.map(function (cell) { return elMap[cell.key]; });
}

/* A component that renders one or more columns of vertical time slots
----------------------------------------------------------------------------------------------------------------------*/
var TimeCols = /** @class */ (function (_super) {
    tslib.__extends(TimeCols, _super);
    function TimeCols() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.processSlotOptions = common.memoize(processSlotOptions);
        _this.state = {
            slatCoords: null,
        };
        _this.handleRootEl = function (el) {
            if (el) {
                _this.context.registerInteractiveComponent(_this, {
                    el: el,
                    isHitComboAllowed: _this.props.isHitComboAllowed,
                });
            }
            else {
                _this.context.unregisterInteractiveComponent(_this);
            }
        };
        _this.handleScrollRequest = function (request) {
            var onScrollTopRequest = _this.props.onScrollTopRequest;
            var slatCoords = _this.state.slatCoords;
            if (onScrollTopRequest && slatCoords) {
                if (request.time) {
                    var top_1 = slatCoords.computeTimeTop(request.time);
                    top_1 = Math.ceil(top_1); // zoom can give weird floating-point values. rather scroll a little bit further
                    if (top_1) {
                        top_1 += 1; // to overcome top border that slots beyond the first have. looks better
                    }
                    onScrollTopRequest(top_1);
                }
                return true;
            }
            return false;
        };
        _this.handleColCoords = function (colCoords) {
            _this.colCoords = colCoords;
        };
        _this.handleSlatCoords = function (slatCoords) {
            _this.setState({ slatCoords: slatCoords });
            if (_this.props.onSlatCoords) {
                _this.props.onSlatCoords(slatCoords);
            }
        };
        return _this;
    }
    TimeCols.prototype.render = function () {
        var _a = this, props = _a.props, state = _a.state;
        return (common.createElement("div", { className: "fc-timegrid-body", ref: this.handleRootEl, style: {
                // these props are important to give this wrapper correct dimensions for interactions
                // TODO: if we set it here, can we avoid giving to inner tables?
                width: props.clientWidth,
                minWidth: props.tableMinWidth,
            } },
            common.createElement(TimeColsSlats, { axis: props.axis, dateProfile: props.dateProfile, slatMetas: props.slatMetas, clientWidth: props.clientWidth, minHeight: props.expandRows ? props.clientHeight : '', tableMinWidth: props.tableMinWidth, tableColGroupNode: props.axis ? props.tableColGroupNode : null /* axis depends on the colgroup's shrinking */, onCoords: this.handleSlatCoords }),
            common.createElement(TimeColsContent, { cells: props.cells, axis: props.axis, dateProfile: props.dateProfile, businessHourSegs: props.businessHourSegs, bgEventSegs: props.bgEventSegs, fgEventSegs: props.fgEventSegs, dateSelectionSegs: props.dateSelectionSegs, eventSelection: props.eventSelection, eventDrag: props.eventDrag, eventResize: props.eventResize, todayRange: props.todayRange, nowDate: props.nowDate, nowIndicatorSegs: props.nowIndicatorSegs, clientWidth: props.clientWidth, tableMinWidth: props.tableMinWidth, tableColGroupNode: props.tableColGroupNode, slatCoords: state.slatCoords, onColCoords: this.handleColCoords, forPrint: props.forPrint })));
    };
    TimeCols.prototype.componentDidMount = function () {
        this.scrollResponder = this.context.createScrollResponder(this.handleScrollRequest);
    };
    TimeCols.prototype.componentDidUpdate = function (prevProps) {
        this.scrollResponder.update(prevProps.dateProfile !== this.props.dateProfile);
    };
    TimeCols.prototype.componentWillUnmount = function () {
        this.scrollResponder.detach();
    };
    TimeCols.prototype.queryHit = function (positionLeft, positionTop) {
        var _a = this.context, dateEnv = _a.dateEnv, options = _a.options;
        var colCoords = this.colCoords;
        var dateProfile = this.props.dateProfile;
        var slatCoords = this.state.slatCoords;
        var _b = this.processSlotOptions(this.props.slotDuration, options.snapDuration), snapDuration = _b.snapDuration, snapsPerSlot = _b.snapsPerSlot;
        var colIndex = colCoords.leftToIndex(positionLeft);
        var slatIndex = slatCoords.positions.topToIndex(positionTop);
        if (colIndex != null && slatIndex != null) {
            var cell = this.props.cells[colIndex];
            var slatTop = slatCoords.positions.tops[slatIndex];
            var slatHeight = slatCoords.positions.getHeight(slatIndex);
            var partial = (positionTop - slatTop) / slatHeight; // floating point number between 0 and 1
            var localSnapIndex = Math.floor(partial * snapsPerSlot); // the snap # relative to start of slat
            var snapIndex = slatIndex * snapsPerSlot + localSnapIndex;
            var dayDate = this.props.cells[colIndex].date;
            var time = common.addDurations(dateProfile.slotMinTime, common.multiplyDuration(snapDuration, snapIndex));
            var start = dateEnv.add(dayDate, time);
            var end = dateEnv.add(start, snapDuration);
            return {
                dateProfile: dateProfile,
                dateSpan: tslib.__assign({ range: { start: start, end: end }, allDay: false }, cell.extraDateSpan),
                dayEl: colCoords.els[colIndex],
                rect: {
                    left: colCoords.lefts[colIndex],
                    right: colCoords.rights[colIndex],
                    top: slatTop,
                    bottom: slatTop + slatHeight,
                },
                layer: 0,
            };
        }
        return null;
    };
    return TimeCols;
}(common.DateComponent));
function processSlotOptions(slotDuration, snapDurationOverride) {
    var snapDuration = snapDurationOverride || slotDuration;
    var snapsPerSlot = common.wholeDivideDurations(slotDuration, snapDuration);
    if (snapsPerSlot === null) {
        snapDuration = slotDuration;
        snapsPerSlot = 1;
        // TODO: say warning?
    }
    return { snapDuration: snapDuration, snapsPerSlot: snapsPerSlot };
}

var DayTimeColsSlicer = /** @class */ (function (_super) {
    tslib.__extends(DayTimeColsSlicer, _super);
    function DayTimeColsSlicer() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    DayTimeColsSlicer.prototype.sliceRange = function (range, dayRanges) {
        var segs = [];
        for (var col = 0; col < dayRanges.length; col += 1) {
            var segRange = common.intersectRanges(range, dayRanges[col]);
            if (segRange) {
                segs.push({
                    start: segRange.start,
                    end: segRange.end,
                    isStart: segRange.start.valueOf() === range.start.valueOf(),
                    isEnd: segRange.end.valueOf() === range.end.valueOf(),
                    col: col,
                });
            }
        }
        return segs;
    };
    return DayTimeColsSlicer;
}(common.Slicer));

var DayTimeCols = /** @class */ (function (_super) {
    tslib.__extends(DayTimeCols, _super);
    function DayTimeCols() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.buildDayRanges = common.memoize(buildDayRanges);
        _this.slicer = new DayTimeColsSlicer();
        _this.timeColsRef = common.createRef();
        return _this;
    }
    DayTimeCols.prototype.render = function () {
        var _this = this;
        var _a = this, props = _a.props, context = _a.context;
        var dateProfile = props.dateProfile, dayTableModel = props.dayTableModel;
        var isNowIndicator = context.options.nowIndicator;
        var dayRanges = this.buildDayRanges(dayTableModel, dateProfile, context.dateEnv);
        // give it the first row of cells
        // TODO: would move this further down hierarchy, but sliceNowDate needs it
        return (common.createElement(common.NowTimer, { unit: isNowIndicator ? 'minute' : 'day' }, function (nowDate, todayRange) { return (common.createElement(TimeCols, tslib.__assign({ ref: _this.timeColsRef }, _this.slicer.sliceProps(props, dateProfile, null, context, dayRanges), { forPrint: props.forPrint, axis: props.axis, dateProfile: dateProfile, slatMetas: props.slatMetas, slotDuration: props.slotDuration, cells: dayTableModel.cells[0], tableColGroupNode: props.tableColGroupNode, tableMinWidth: props.tableMinWidth, clientWidth: props.clientWidth, clientHeight: props.clientHeight, expandRows: props.expandRows, nowDate: nowDate, nowIndicatorSegs: isNowIndicator && _this.slicer.sliceNowDate(nowDate, context, dayRanges), todayRange: todayRange, onScrollTopRequest: props.onScrollTopRequest, onSlatCoords: props.onSlatCoords }))); }));
    };
    return DayTimeCols;
}(common.DateComponent));
function buildDayRanges(dayTableModel, dateProfile, dateEnv) {
    var ranges = [];
    for (var _i = 0, _a = dayTableModel.headerDates; _i < _a.length; _i++) {
        var date = _a[_i];
        ranges.push({
            start: dateEnv.add(date, dateProfile.slotMinTime),
            end: dateEnv.add(date, dateProfile.slotMaxTime),
        });
    }
    return ranges;
}

// potential nice values for the slot-duration and interval-duration
// from largest to smallest
var STOCK_SUB_DURATIONS = [
    { hours: 1 },
    { minutes: 30 },
    { minutes: 15 },
    { seconds: 30 },
    { seconds: 15 },
];
function buildSlatMetas(slotMinTime, slotMaxTime, explicitLabelInterval, slotDuration, dateEnv) {
    var dayStart = new Date(0);
    var slatTime = slotMinTime;
    var slatIterator = common.createDuration(0);
    var labelInterval = explicitLabelInterval || computeLabelInterval(slotDuration);
    var metas = [];
    while (common.asRoughMs(slatTime) < common.asRoughMs(slotMaxTime)) {
        var date = dateEnv.add(dayStart, slatTime);
        var isLabeled = common.wholeDivideDurations(slatIterator, labelInterval) !== null;
        metas.push({
            date: date,
            time: slatTime,
            key: date.toISOString(),
            isoTimeStr: common.formatIsoTimeString(date),
            isLabeled: isLabeled,
        });
        slatTime = common.addDurations(slatTime, slotDuration);
        slatIterator = common.addDurations(slatIterator, slotDuration);
    }
    return metas;
}
// Computes an automatic value for slotLabelInterval
function computeLabelInterval(slotDuration) {
    var i;
    var labelInterval;
    var slotsPerLabel;
    // find the smallest stock label interval that results in more than one slots-per-label
    for (i = STOCK_SUB_DURATIONS.length - 1; i >= 0; i -= 1) {
        labelInterval = common.createDuration(STOCK_SUB_DURATIONS[i]);
        slotsPerLabel = common.wholeDivideDurations(labelInterval, slotDuration);
        if (slotsPerLabel !== null && slotsPerLabel > 1) {
            return labelInterval;
        }
    }
    return slotDuration; // fall back
}

var DayTimeColsView = /** @class */ (function (_super) {
    tslib.__extends(DayTimeColsView, _super);
    function DayTimeColsView() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.buildTimeColsModel = common.memoize(buildTimeColsModel);
        _this.buildSlatMetas = common.memoize(buildSlatMetas);
        return _this;
    }
    DayTimeColsView.prototype.render = function () {
        var _this = this;
        var _a = this.context, options = _a.options, dateEnv = _a.dateEnv, dateProfileGenerator = _a.dateProfileGenerator;
        var props = this.props;
        var dateProfile = props.dateProfile;
        var dayTableModel = this.buildTimeColsModel(dateProfile, dateProfileGenerator);
        var splitProps = this.allDaySplitter.splitProps(props);
        var slatMetas = this.buildSlatMetas(dateProfile.slotMinTime, dateProfile.slotMaxTime, options.slotLabelInterval, options.slotDuration, dateEnv);
        var dayMinWidth = options.dayMinWidth;
        var hasAttachedAxis = !dayMinWidth;
        var hasDetachedAxis = dayMinWidth;
        var headerContent = options.dayHeaders && (common.createElement(common.DayHeader, { dates: dayTableModel.headerDates, dateProfile: dateProfile, datesRepDistinctDays: true, renderIntro: hasAttachedAxis ? this.renderHeadAxis : null }));
        var allDayContent = (options.allDaySlot !== false) && (function (contentArg) { return (common.createElement(daygrid.DayTable, tslib.__assign({}, splitProps.allDay, { dateProfile: dateProfile, dayTableModel: dayTableModel, nextDayThreshold: options.nextDayThreshold, tableMinWidth: contentArg.tableMinWidth, colGroupNode: contentArg.tableColGroupNode, renderRowIntro: hasAttachedAxis ? _this.renderTableRowAxis : null, showWeekNumbers: false, expandRows: false, headerAlignElRef: _this.headerElRef, clientWidth: contentArg.clientWidth, clientHeight: contentArg.clientHeight, forPrint: props.forPrint }, _this.getAllDayMaxEventProps()))); });
        var timeGridContent = function (contentArg) { return (common.createElement(DayTimeCols, tslib.__assign({}, splitProps.timed, { dayTableModel: dayTableModel, dateProfile: dateProfile, axis: hasAttachedAxis, slotDuration: options.slotDuration, slatMetas: slatMetas, forPrint: props.forPrint, tableColGroupNode: contentArg.tableColGroupNode, tableMinWidth: contentArg.tableMinWidth, clientWidth: contentArg.clientWidth, clientHeight: contentArg.clientHeight, onSlatCoords: _this.handleSlatCoords, expandRows: contentArg.expandRows, onScrollTopRequest: _this.handleScrollTopRequest }))); };
        return hasDetachedAxis
            ? this.renderHScrollLayout(headerContent, allDayContent, timeGridContent, dayTableModel.colCnt, dayMinWidth, slatMetas, this.state.slatCoords)
            : this.renderSimpleLayout(headerContent, allDayContent, timeGridContent);
    };
    return DayTimeColsView;
}(TimeColsView));
function buildTimeColsModel(dateProfile, dateProfileGenerator) {
    var daySeries = new common.DaySeriesModel(dateProfile.renderRange, dateProfileGenerator);
    return new common.DayTableModel(daySeries, false);
}

var OPTION_REFINERS = {
    allDaySlot: Boolean,
};

var main = common.createPlugin({
    initialView: 'timeGridWeek',
    optionRefiners: OPTION_REFINERS,
    views: {
        timeGrid: {
            component: DayTimeColsView,
            usesMinMaxTime: true,
            allDaySlot: true,
            slotDuration: '00:30:00',
            slotEventOverlap: true, // a bad name. confused with overlap/constraint system
        },
        timeGridDay: {
            type: 'timeGrid',
            duration: { days: 1 },
        },
        timeGridWeek: {
            type: 'timeGrid',
            duration: { weeks: 1 },
        },
    },
});

exports.DayTimeCols = DayTimeCols;
exports.DayTimeColsSlicer = DayTimeColsSlicer;
exports.DayTimeColsView = DayTimeColsView;
exports.TimeCols = TimeCols;
exports.TimeColsSlatsCoords = TimeColsSlatsCoords;
exports.TimeColsView = TimeColsView;
exports.buildDayRanges = buildDayRanges;
exports.buildSlatMetas = buildSlatMetas;
exports.buildTimeColsModel = buildTimeColsModel;
exports.default = main;
