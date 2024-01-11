use std::sync::Once;

use env_logger::Env;

static INIT: Once = Once::new();

pub fn init() {
    INIT.call_once(|| {
        let _ = env_logger::Builder::from_env(Env::default().default_filter_or("debug"))
            .is_test(true)
            .try_init();
    });
}
