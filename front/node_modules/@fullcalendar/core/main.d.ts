import './vdom';
import * as _fullcalendar_common from '@fullcalendar/common';
import { CalendarApi, CalendarData, DelayedRunner, CalendarOptions, Action, CssDimValue } from '@fullcalendar/common';
export * from '@fullcalendar/common';

declare class Calendar extends CalendarApi {
    currentData: CalendarData;
    renderRunner: DelayedRunner;
    el: HTMLElement;
    isRendering: boolean;
    isRendered: boolean;
    currentClassNames: string[];
    customContentRenderId: number;
    get view(): _fullcalendar_common.ViewApi;
    constructor(el: HTMLElement, optionOverrides?: CalendarOptions);
    handleAction: (action: Action) => void;
    handleData: (data: CalendarData) => void;
    handleRenderRequest: () => void;
    render(): void;
    destroy(): void;
    updateSize(): void;
    batchRendering(func: any): void;
    pauseRendering(): void;
    resumeRendering(): void;
    resetOptions(optionOverrides: any, append?: any): void;
    setClassNames(classNames: string[]): void;
    setHeight(height: CssDimValue): void;
}

export { Calendar };
