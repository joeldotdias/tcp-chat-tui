use std::io;

use app::App;
use crossterm::terminal::{
    disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen,
};
use futures::{SinkExt, StreamExt};
use ratatui::{
    backend::CrosstermBackend,
    layout::{Constraint, Layout},
    style::{Color, Modifier, Style, Stylize},
    text::Span,
    widgets::{Block, Borders},
    Terminal,
};
use tokio::net::TcpStream;
use tokio_util::codec::{FramedRead, FramedWrite, LinesCodec};
use tui_textarea::{Input, Key, TextArea};

mod app;

#[tokio::main]
async fn main() -> io::Result<()> {
    let mut tcp_conn = TcpStream::connect("127.0.0.1:6969").await?;
    let (reader, writer) = tcp_conn.split();
    let mut tcp_sink = FramedWrite::new(writer, LinesCodec::new());
    let mut tcp_stream = FramedRead::new(reader, LinesCodec::new());

    let stdout = io::stdout();
    let mut stdout = stdout.lock();
    let mut app = App::new();

    enable_raw_mode()?;
    crossterm::execute!(stdout, EnterAlternateScreen)?;
    let mut term = Terminal::new(CrosstermBackend::new(stdout))?;

    let layout = Layout::vertical([Constraint::Percentage(100), Constraint::Min(3)]);
    let mut text_area = text_area_refresh();

    let mut term_stream = crossterm::event::EventStream::new();

    loop {
        term.draw(|f| {
            let chunks = layout.split(f.size());
            let messages_area = chunks[0];
            let input_area = chunks[1];

            let msg_height = messages_area.height - 2;
            let msg_width = messages_area.width - 2;

            let msgeez = app
                .messages_to_list(msg_height.into(), msg_width.into())
                .block(
                    Block::default()
                        .rapid_blink()
                        .borders(Borders::ALL)
                        .title(Span::styled(
                            &app.room,
                            Style::default().fg(Color::Red).add_modifier(Modifier::BOLD),
                        ))
                        .title_bottom(
                            Span::styled(
                                "Type \":help\" for help",
                                Style::default().fg(Color::Gray).add_modifier(Modifier::DIM),
                            )
                            .into_right_aligned_line(),
                        ),
                );

            f.render_widget(msgeez, messages_area);

            let input = text_area.widget();
            f.render_widget(input, input_area);
        })?;

        tokio::select! {
            term_event = term_stream.next() => {
                if let Some(event) = term_event {
                    let event = match event {
                        Ok(event) => event,
                        Err(_) => break,
                    };
                    match event.into() {
                        Input { key: Key::Char('c'), ctrl: true, .. } => break,
                        Input { key: Key::Enter, .. }=> {
                            if text_area.is_empty() {
                                continue;
                            }

                            for line in text_area.clone().into_lines() {
                                match tcp_sink.send(line).await {
                                    Ok(_) => (),
                                    Err(_) => break,
                                }
                            }
                            text_area = text_area_refresh();
                        }
                        input => {
                            text_area.input_without_shortcuts(input);
                        }
                    }
                } else {
                    break;
                }
            }

            tcp_event = tcp_stream.next() => match tcp_event {
                Some(event) => {
                    let server_msg = match event {
                        Ok(msg) => msg,
                        Err(_) => break,
                    };

                    if server_msg.starts_with("*** Welcome to") {
                        let room_name = server_msg
                            .split_ascii_whitespace()
                            .nth(3)
                            .unwrap();

                        app.update_room_name(room_name);
                    } else if server_msg.starts_with("*** You are now") {
                        let _name = server_msg
                            .split_ascii_whitespace()
                            .nth(4)
                            .unwrap();
                    }
                    app.add_msg(server_msg);
                }
                None => break,
            }
        };
    }

    disable_raw_mode()?;
    crossterm::execute!(term.backend_mut(), LeaveAlternateScreen)?;
    term.show_cursor()?;

    Ok(())
}

fn text_area_refresh() -> TextArea<'static> {
    let mut text_area = TextArea::default();
    text_area.set_cursor_line_style(Style::default().add_modifier(Modifier::SLOW_BLINK));
    text_area.set_placeholder_text("Message...");
    text_area.set_block(Block::default().borders(Borders::ALL).title("Send"));

    text_area
}
