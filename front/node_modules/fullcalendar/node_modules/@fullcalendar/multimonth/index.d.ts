import { PluginDef } from '@fullcalendar/core';
import { createFormatter } from '@fullcalendar/core/internal';

declare const OPTION_REFINERS: {
    multiMonthTitleFormat: typeof createFormatter;
    multiMonthMaxColumns: NumberConstructor;
    multiMonthMinWidth: NumberConstructor;
};

type ExtraOptionRefiners = typeof OPTION_REFINERS;
declare module '@fullcalendar/core/internal' {
    interface BaseOptionRefiners extends ExtraOptionRefiners {
    }
}
//# sourceMappingURL=ambient.d.ts.map

declare const _default: PluginDef;
//# sourceMappingURL=index.d.ts.map

export { _default as default };
