import { createPlugin } from '@fullcalendar/core/index.js';
import { buildDayTableModel, DayTableSlicer, TableRows, buildDayTableRenderRange, TableDateProfileGenerator } from '@fullcalendar/daygrid/internal.js';
import { DateComponent, memoize, getUniqueDomId, DayHeader, ViewContainer, formatIsoMonthStr, isPropsEqual, createDuration, intersectRanges, createFormatter, injectStyles } from '@fullcalendar/core/internal.js';
import { createElement, createRef } from '@fullcalendar/core/preact.js';

class SingleMonth extends DateComponent {
    constructor() {
        super(...arguments);
        this.buildDayTableModel = memoize(buildDayTableModel);
        this.slicer = new DayTableSlicer();
        this.state = {
            labelId: getUniqueDomId(),
        };
    }
    render() {
        const { props, state, context } = this;
        const { dateProfile, forPrint } = props;
        const { options } = context;
        const dayTableModel = this.buildDayTableModel(dateProfile, context.dateProfileGenerator);
        const slicedProps = this.slicer.sliceProps(props, dateProfile, options.nextDayThreshold, context, dayTableModel);
        // ensure single-month has aspect ratio
        const tableHeight = props.tableWidth != null ? props.tableWidth / options.aspectRatio : null;
        const rowCnt = dayTableModel.cells.length;
        const rowHeight = tableHeight != null ? tableHeight / rowCnt : null;
        return (createElement("div", { ref: props.elRef, "data-date": props.isoDateStr, className: "fc-multimonth-month", style: { width: props.width }, role: "grid", "aria-labelledby": state.labelId },
            createElement("div", { className: "fc-multimonth-header", style: { marginBottom: rowHeight }, role: "presentation" },
                createElement("div", { className: "fc-multimonth-title", id: state.labelId }, context.dateEnv.format(props.dateProfile.currentRange.start, props.titleFormat)),
                createElement("table", { className: [
                        'fc-multimonth-header-table',
                        context.theme.getClass('table'),
                    ].join(' '), role: "presentation" },
                    createElement("thead", { role: "rowgroup" },
                        createElement(DayHeader, { dateProfile: props.dateProfile, dates: dayTableModel.headerDates, datesRepDistinctDays: false })))),
            createElement("div", { className: [
                    'fc-multimonth-daygrid',
                    'fc-daygrid',
                    'fc-daygrid-body',
                    !forPrint && 'fc-daygrid-body-balanced',
                    forPrint && 'fc-daygrid-body-unbalanced',
                    forPrint && 'fc-daygrid-body-natural',
                ].join(' '), style: { marginTop: -rowHeight } },
                createElement("table", { className: [
                        'fc-multimonth-daygrid-table',
                        context.theme.getClass('table'),
                    ].join(' '), style: { height: forPrint ? '' : tableHeight }, role: "presentation" },
                    createElement("tbody", { role: "rowgroup" },
                        createElement(TableRows, Object.assign({}, slicedProps, { dateProfile: dateProfile, cells: dayTableModel.cells, eventSelection: props.eventSelection, dayMaxEvents: !forPrint, dayMaxEventRows: !forPrint, showWeekNumbers: options.weekNumbers, clientWidth: props.clientWidth, clientHeight: props.clientHeight, forPrint: forPrint })))))));
    }
}

class MultiMonthView extends DateComponent {
    constructor() {
        super(...arguments);
        this.splitDateProfileByMonth = memoize(splitDateProfileByMonth);
        this.buildMonthFormat = memoize(buildMonthFormat);
        this.scrollElRef = createRef();
        this.firstMonthElRef = createRef();
        this.needsScrollReset = false;
        this.handleSizing = (isForced) => {
            if (isForced) {
                this.updateSize();
            }
        };
    }
    render() {
        const { context, props, state } = this;
        const { options } = context;
        const { clientWidth, clientHeight } = state;
        const monthHPadding = state.monthHPadding || 0;
        const colCount = Math.min(clientWidth != null ?
            Math.floor(clientWidth / (options.multiMonthMinWidth + monthHPadding)) :
            1, options.multiMonthMaxColumns) || 1;
        const monthWidthPct = (100 / colCount) + '%';
        const monthTableWidth = clientWidth == null ? null :
            (clientWidth / colCount) - monthHPadding;
        const isLegitSingleCol = clientWidth != null && colCount === 1;
        const monthDateProfiles = this.splitDateProfileByMonth(context.dateProfileGenerator, props.dateProfile, context.dateEnv, isLegitSingleCol ? false : options.fixedWeekCount, options.showNonCurrentDates);
        const monthTitleFormat = this.buildMonthFormat(options.multiMonthTitleFormat, monthDateProfiles);
        const rootClassNames = [
            'fc-multimonth',
            isLegitSingleCol ?
                'fc-multimonth-singlecol' :
                'fc-multimonth-multicol',
            (monthTableWidth != null && monthTableWidth < 400) ?
                'fc-multimonth-compact' :
                '',
        ];
        return (createElement(ViewContainer, { elRef: this.scrollElRef, elClasses: rootClassNames, viewSpec: context.viewSpec }, monthDateProfiles.map((monthDateProfile, i) => {
            const monthStr = formatIsoMonthStr(monthDateProfile.currentRange.start);
            return (createElement(SingleMonth, Object.assign({}, props, { key: monthStr, isoDateStr: monthStr, elRef: i === 0 ? this.firstMonthElRef : undefined, titleFormat: monthTitleFormat, dateProfile: monthDateProfile, width: monthWidthPct, tableWidth: monthTableWidth, clientWidth: clientWidth, clientHeight: clientHeight })));
        })));
    }
    componentDidMount() {
        this.updateSize();
        this.context.addResizeHandler(this.handleSizing);
        this.requestScrollReset();
    }
    componentDidUpdate(prevProps) {
        if (!isPropsEqual(prevProps, this.props)) { // an external change?
            this.handleSizing(false);
        }
        if (prevProps.dateProfile !== this.props.dateProfile) {
            this.requestScrollReset();
        }
        else {
            this.flushScrollReset();
        }
    }
    componentWillUnmount() {
        this.context.removeResizeHandler(this.handleSizing);
    }
    updateSize() {
        const scrollEl = this.scrollElRef.current;
        const firstMonthEl = this.firstMonthElRef.current;
        if (scrollEl) {
            this.setState({
                clientWidth: scrollEl.clientWidth,
                clientHeight: scrollEl.clientHeight,
            });
        }
        if (firstMonthEl && scrollEl) {
            if (this.state.monthHPadding == null) { // always remember initial non-zero value
                this.setState({
                    monthHPadding: scrollEl.clientWidth - // go within padding
                        firstMonthEl.firstChild.offsetWidth,
                });
            }
        }
    }
    requestScrollReset() {
        this.needsScrollReset = true;
        this.flushScrollReset();
    }
    flushScrollReset() {
        if (this.needsScrollReset &&
            this.state.monthHPadding != null // indicates sizing already happened
        ) {
            const { currentDate } = this.props.dateProfile;
            const scrollEl = this.scrollElRef.current;
            const monthEl = scrollEl.querySelector(`[data-date="${formatIsoMonthStr(currentDate)}"]`);
            scrollEl.scrollTop = monthEl.getBoundingClientRect().top -
                this.firstMonthElRef.current.getBoundingClientRect().top;
            this.needsScrollReset = false;
        }
    }
    // workaround for when queued setState render (w/ clientWidth) gets cancelled because
    // subsequent update and shouldComponentUpdate says not to render :(
    shouldComponentUpdate() {
        return true;
    }
}
// date profile
// -------------------------------------------------------------------------------------------------
const oneMonthDuration = createDuration(1, 'month');
function splitDateProfileByMonth(dateProfileGenerator, dateProfile, dateEnv, fixedWeekCount, showNonCurrentDates) {
    const { start, end } = dateProfile.currentRange;
    let monthStart = start;
    const monthDateProfiles = [];
    while (monthStart.valueOf() < end.valueOf()) {
        const monthEnd = dateEnv.add(monthStart, oneMonthDuration);
        const currentRange = {
            // yuck
            start: dateProfileGenerator.skipHiddenDays(monthStart),
            end: dateProfileGenerator.skipHiddenDays(monthEnd, -1, true),
        };
        let renderRange = buildDayTableRenderRange({
            currentRange,
            snapToWeek: true,
            fixedWeekCount,
            dateEnv,
        });
        renderRange = {
            // yuck
            start: dateProfileGenerator.skipHiddenDays(renderRange.start),
            end: dateProfileGenerator.skipHiddenDays(renderRange.end, -1, true),
        };
        const activeRange = dateProfile.activeRange ?
            intersectRanges(dateProfile.activeRange, showNonCurrentDates ? renderRange : currentRange) :
            null;
        monthDateProfiles.push({
            currentDate: dateProfile.currentDate,
            isValid: dateProfile.isValid,
            validRange: dateProfile.validRange,
            renderRange,
            activeRange,
            currentRange,
            currentRangeUnit: 'month',
            isRangeAllDay: true,
            dateIncrement: dateProfile.dateIncrement,
            slotMinTime: dateProfile.slotMaxTime,
            slotMaxTime: dateProfile.slotMinTime,
        });
        monthStart = monthEnd;
    }
    return monthDateProfiles;
}
// date formatting
// -------------------------------------------------------------------------------------------------
const YEAR_MONTH_FORMATTER = createFormatter({ year: 'numeric', month: 'long' });
const YEAR_FORMATTER = createFormatter({ month: 'long' });
function buildMonthFormat(formatOverride, monthDateProfiles) {
    return formatOverride ||
        ((monthDateProfiles[0].currentRange.start.getUTCFullYear() !==
            monthDateProfiles[monthDateProfiles.length - 1].currentRange.start.getUTCFullYear())
            ? YEAR_MONTH_FORMATTER
            : YEAR_FORMATTER);
}

const OPTION_REFINERS = {
    multiMonthTitleFormat: createFormatter,
    multiMonthMaxColumns: Number,
    multiMonthMinWidth: Number,
};

var css_248z = ".fc .fc-multimonth{border:1px solid var(--fc-border-color);display:flex;flex-wrap:wrap;overflow-x:hidden;overflow-y:auto}.fc .fc-multimonth-title{font-size:1.2em;font-weight:700;padding:1em 0;text-align:center}.fc .fc-multimonth-daygrid{background:var(--fc-page-bg-color)}.fc .fc-multimonth-daygrid-table,.fc .fc-multimonth-header-table{table-layout:fixed;width:100%}.fc .fc-multimonth-daygrid-table{border-top-style:hidden!important}.fc .fc-multimonth-singlecol .fc-multimonth{position:relative}.fc .fc-multimonth-singlecol .fc-multimonth-header{background:var(--fc-page-bg-color);position:relative;top:0;z-index:2}.fc .fc-multimonth-singlecol .fc-multimonth-daygrid{position:relative;z-index:1}.fc .fc-multimonth-singlecol .fc-multimonth-daygrid-table,.fc .fc-multimonth-singlecol .fc-multimonth-header-table{border-left-style:hidden;border-right-style:hidden}.fc .fc-multimonth-singlecol .fc-multimonth-month:last-child .fc-multimonth-daygrid-table{border-bottom-style:hidden}.fc .fc-multimonth-multicol{line-height:1}.fc .fc-multimonth-multicol .fc-multimonth-month{padding:0 1.2em 1.2em}.fc .fc-multimonth-multicol .fc-daygrid-more-link{border:1px solid var(--fc-event-border-color);display:block;float:none;padding:1px}.fc .fc-multimonth-compact{line-height:1}.fc .fc-multimonth-compact .fc-multimonth-daygrid-table,.fc .fc-multimonth-compact .fc-multimonth-header-table{font-size:.9em}.fc-media-screen .fc-multimonth-singlecol .fc-multimonth-header{position:sticky}.fc-media-print .fc-multimonth{overflow:visible}";
injectStyles(css_248z);

var index = createPlugin({
    name: '@fullcalendar/multimonth',
    initialView: 'multiMonthYear',
    optionRefiners: OPTION_REFINERS,
    views: {
        multiMonth: {
            component: MultiMonthView,
            dateProfileGeneratorClass: TableDateProfileGenerator,
            multiMonthMinWidth: 350,
            multiMonthMaxColumns: 3,
        },
        multiMonthYear: {
            type: 'multiMonth',
            duration: { years: 1 },
            fixedWeekCount: true,
            showNonCurrentDates: false,
        },
    },
});

export { index as default };
