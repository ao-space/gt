use std::pin::Pin;
use std::task::Poll;

use hyper::rt;
use tokio::io;

pub struct Compat<T>(T);

impl<T> Compat<T> {
    pub fn new(io: T) -> Self {
        Compat(io)
    }

    fn p(self: Pin<&mut Self>) -> Pin<&mut T> {
        unsafe { self.map_unchecked_mut(|me| &mut me.0) }
    }
}

impl<T> rt::Read for Compat<T>
where
    T: io::AsyncRead,
{
    fn poll_read(
        self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        mut buf: rt::ReadBufCursor<'_>,
    ) -> Poll<Result<(), io::Error>> {
        let n = unsafe {
            let mut tbuf = io::ReadBuf::uninit(buf.as_mut());
            match io::AsyncRead::poll_read(self.p(), cx, &mut tbuf) {
                Poll::Ready(Ok(())) => tbuf.filled().len(),
                other => return other,
            }
        };

        unsafe {
            buf.advance(n);
        }
        Poll::Ready(Ok(()))
    }
}

impl<T> rt::Write for Compat<T>
where
    T: io::AsyncWrite,
{
    fn poll_write(
        self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        buf: &[u8],
    ) -> Poll<Result<usize, io::Error>> {
        io::AsyncWrite::poll_write(self.p(), cx, buf)
    }

    fn poll_flush(
        self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> Poll<Result<(), io::Error>> {
        io::AsyncWrite::poll_flush(self.p(), cx)
    }

    fn poll_shutdown(
        self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
    ) -> Poll<Result<(), io::Error>> {
        io::AsyncWrite::poll_shutdown(self.p(), cx)
    }

    fn is_write_vectored(&self) -> bool {
        io::AsyncWrite::is_write_vectored(&self.0)
    }

    fn poll_write_vectored(
        self: Pin<&mut Self>,
        cx: &mut std::task::Context<'_>,
        bufs: &[std::io::IoSlice<'_>],
    ) -> Poll<Result<usize, io::Error>> {
        io::AsyncWrite::poll_write_vectored(self.p(), cx, bufs)
    }
}
