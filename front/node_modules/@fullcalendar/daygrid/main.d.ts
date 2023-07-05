
import * as _fullcalendar_common from '@fullcalendar/common';
import { Dictionary, DateComponent, ViewProps, RefObject, ChunkConfigRowContent, ChunkContentCallbackArgs, VNode, createElement, DateProfile, DateProfileGenerator, DayTableModel, ViewContext, Duration, EventStore, EventUiHash, DateSpan, EventInteractionState, CssDimValue, Seg, Slicer, DateRange, Hit, DayTableCell, EventSegUiInteractionState } from '@fullcalendar/common';

declare abstract class TableView<State = Dictionary> extends DateComponent<ViewProps, State> {
    protected headerElRef: RefObject<HTMLTableCellElement>;
    renderSimpleLayout(headerRowContent: ChunkConfigRowContent, bodyContent: (contentArg: ChunkContentCallbackArgs) => VNode): createElement.JSX.Element;
    renderHScrollLayout(headerRowContent: ChunkConfigRowContent, bodyContent: (contentArg: ChunkContentCallbackArgs) => VNode, colCnt: number, dayMinWidth: number): createElement.JSX.Element;
}

declare class DayTableView extends TableView {
    private buildDayTableModel;
    private headerRef;
    private tableRef;
    render(): createElement.JSX.Element;
}
declare function buildDayTableModel(dateProfile: DateProfile, dateProfileGenerator: DateProfileGenerator): DayTableModel;

interface DayTableProps {
    dateProfile: DateProfile;
    dayTableModel: DayTableModel;
    nextDayThreshold: Duration;
    businessHours: EventStore;
    eventStore: EventStore;
    eventUiBases: EventUiHash;
    dateSelection: DateSpan | null;
    eventSelection: string;
    eventDrag: EventInteractionState | null;
    eventResize: EventInteractionState | null;
    colGroupNode: VNode;
    tableMinWidth: CssDimValue;
    renderRowIntro?: () => VNode;
    dayMaxEvents: boolean | number;
    dayMaxEventRows: boolean | number;
    expandRows: boolean;
    showWeekNumbers: boolean;
    headerAlignElRef?: RefObject<HTMLElement>;
    clientWidth: number | null;
    clientHeight: number | null;
    forPrint: boolean;
}
declare class DayTable extends DateComponent<DayTableProps, ViewContext> {
    private slicer;
    private tableRef;
    render(): createElement.JSX.Element;
}

interface TableSeg extends Seg {
    row: number;
    firstCol: number;
    lastCol: number;
}

declare class DayTableSlicer extends Slicer<TableSeg, [DayTableModel]> {
    forceDayIfListItem: boolean;
    sliceRange(dateRange: DateRange, dayTableModel: DayTableModel): TableSeg[];
}

interface TableProps {
    dateProfile: DateProfile;
    cells: DayTableCell[][];
    renderRowIntro?: () => VNode;
    colGroupNode: VNode;
    tableMinWidth: CssDimValue;
    expandRows: boolean;
    showWeekNumbers: boolean;
    clientWidth: number | null;
    clientHeight: number | null;
    businessHourSegs: TableSeg[];
    bgEventSegs: TableSeg[];
    fgEventSegs: TableSeg[];
    dateSelectionSegs: TableSeg[];
    eventSelection: string;
    eventDrag: EventSegUiInteractionState | null;
    eventResize: EventSegUiInteractionState | null;
    dayMaxEvents: boolean | number;
    dayMaxEventRows: boolean | number;
    headerAlignElRef?: RefObject<HTMLElement>;
    forPrint: boolean;
    isHitComboAllowed?: (hit0: Hit, hit1: Hit) => boolean;
}
declare class Table extends DateComponent<TableProps> {
    private splitBusinessHourSegs;
    private splitBgEventSegs;
    private splitFgEventSegs;
    private splitDateSelectionSegs;
    private splitEventDrag;
    private splitEventResize;
    private rootEl;
    private rowRefs;
    private rowPositions;
    private colPositions;
    render(): createElement.JSX.Element;
    handleRootEl: (rootEl: HTMLElement | null) => void;
    prepareHits(): void;
    queryHit(positionLeft: number, positionTop: number): Hit;
    private getCellEl;
    private getCellRange;
}

declare const _default: _fullcalendar_common.PluginDef;


export default _default;
export { DayTableView as DayGridView, DayTable, DayTableSlicer, Table, TableSeg, TableView, buildDayTableModel };
