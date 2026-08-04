import { useEffect, useMemo, useState, type JSX } from 'react'
import { Link } from 'react-router-dom'
import styles from './UserGuidePage.module.css'

const WEB = 'https://ecomerce.nvnhan0810.com'
const MERCHANT = 'https://ecomerce-merchant.nvnhan0810.com'
const ADMIN = 'https://ecomerce-admin.nvnhan0810.com'

type TocItem = {
  id: string
  label: string
  children?: { id: string; label: string }[]
}

const TOC: TocItem[] = [
  { id: 'tong-quan', label: 'Tổng quan' },
  { id: 'tai-khoan-demo', label: 'Tài khoản demo' },
  {
    id: 'web-khach-hang',
    label: 'Web khách hàng',
    children: [
      { id: 'web-dang-nhap', label: 'Đăng nhập & tài khoản' },
      { id: 'web-mua-hang', label: 'Mua hàng' },
      { id: 'web-dia-chi', label: 'Địa chỉ giao hàng' },
      { id: 'web-don-hang', label: 'Đơn hàng & thanh toán' },
    ],
  },
  {
    id: 'merchant',
    label: 'Cổng cửa hàng (Merchant)',
    children: [
      { id: 'merchant-dang-nhap', label: 'Đăng nhập & gian hàng' },
      { id: 'merchant-san-pham', label: 'Quản lý sản phẩm' },
      { id: 'merchant-don-hang', label: 'Xử lý đơn hàng' },
    ],
  },
  {
    id: 'admin',
    label: 'Trang quản trị (Admin)',
    children: [
      { id: 'admin-dang-nhap', label: 'Đăng nhập & tổng quan' },
      { id: 'admin-quan-ly', label: 'Người dùng & danh mục' },
      { id: 'admin-don-hang', label: 'Đơn hàng & giao hàng' },
      { id: 'admin-thanh-toan', label: 'Thanh toán online' },
    ],
  },
  { id: 'onepay', label: 'Thẻ test thanh toán' },
  { id: 'luong-demo', label: 'Luồng demo nhanh' },
]

const SECTION_IDS = TOC.flatMap((item) => [
  item.id,
  ...(item.children?.map((c) => c.id) ?? []),
])

function CredTable({
  rows,
}: {
  rows: { role: string; email: string; password: string; note?: string }[]
}): JSX.Element {
  return (
    <div className={styles.tableWrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th>Vai trò / tên</th>
            <th>Email</th>
            <th>Mật khẩu</th>
            <th>Ghi chú</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.email}>
              <td>{row.role}</td>
              <td>
                <code>{row.email}</code>
              </td>
              <td>
                <code>{row.password}</code>
              </td>
              <td>{row.note || '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function FeatureLink({
  href,
  children,
  external,
}: {
  href: string
  children: string
  external?: boolean
}): JSX.Element {
  if (external) {
    return (
      <a className={styles.featureLink} href={href} target="_blank" rel="noreferrer">
        {children}
      </a>
    )
  }
  return (
    <Link className={styles.featureLink} to={href}>
      {children}
    </Link>
  )
}

export function UserGuidePage(): JSX.Element {
  const [activeId, setActiveId] = useState(TOC[0].id)
  const idsKey = useMemo(() => SECTION_IDS.join('|'), [])

  useEffect(() => {
    const nodes = SECTION_IDS.map((id) => document.getElementById(id)).filter(
      (el): el is HTMLElement => Boolean(el),
    )
    if (nodes.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)
        if (visible[0]?.target.id) {
          setActiveId(visible[0].target.id)
        }
      },
      {
        rootMargin: '-20% 0px -55% 0px',
        threshold: [0.1, 0.35, 0.6],
      },
    )
    nodes.forEach((n) => observer.observe(n))
    return () => observer.disconnect()
  }, [idsKey])

  function scrollToSection(id: string) {
    const el = document.getElementById(id)
    if (!el) return
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setActiveId(id)
    window.history.replaceState(null, '', `#${id}`)
  }

  useEffect(() => {
    const hash = window.location.hash.replace(/^#/, '')
    if (hash && SECTION_IDS.includes(hash)) {
      requestAnimationFrame(() => scrollToSection(hash))
    }
  }, [])

  return (
    <div className={styles.page}>
      <aside className={styles.toc} aria-label="Mục lục hướng dẫn">
        <p className={styles.tocTitle}>Mục lục</p>
        <nav className={styles.tocNav}>
          {TOC.map((item) => (
            <div key={item.id} className={styles.tocGroup}>
              <button
                type="button"
                className={`${styles.tocLink} ${activeId === item.id ? styles.tocActive : ''}`}
                onClick={() => scrollToSection(item.id)}
              >
                {item.label}
              </button>
              {item.children ? (
                <div className={styles.tocChildren}>
                  {item.children.map((child) => (
                    <button
                      key={child.id}
                      type="button"
                      className={`${styles.tocSubLink} ${activeId === child.id ? styles.tocActive : ''}`}
                      onClick={() => scrollToSection(child.id)}
                    >
                      {child.label}
                    </button>
                  ))}
                </div>
              ) : null}
            </div>
          ))}
        </nav>
      </aside>

      <article className={styles.content}>
        <header className={styles.hero}>
          <p className={styles.kicker}>Hướng dẫn nghiệp vụ</p>
          <h1>Hướng dẫn sử dụng</h1>
          <p className={styles.lead}>
            Tài liệu mô tả cách dùng 3 ứng dụng demo: web mua hàng, cổng cửa hàng và trang quản trị —
            kèm tài khoản mẫu và các bước thao tác chính.
          </p>
        </header>

        <section id="tong-quan" className={styles.section}>
          <h2>Tổng quan</h2>
          <p>Hệ thống demo gồm 3 ứng dụng, mỗi ứng dụng phục vụ một vai trò:</p>
          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Ứng dụng</th>
                  <th>Đường dẫn</th>
                  <th>Dành cho</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>Web khách hàng</td>
                  <td>
                    <a href={WEB} target="_blank" rel="noreferrer">
                      {WEB}
                    </a>
                  </td>
                  <td>Khách mua hàng, theo dõi đơn, thanh toán</td>
                </tr>
                <tr>
                  <td>Cổng cửa hàng</td>
                  <td>
                    <a href={MERCHANT} target="_blank" rel="noreferrer">
                      {MERCHANT}
                    </a>
                  </td>
                  <td>Chủ shop quản lý sản phẩm và đơn hàng</td>
                </tr>
                <tr>
                  <td>Trang quản trị</td>
                  <td>
                    <a href={ADMIN} target="_blank" rel="noreferrer">
                      {ADMIN}
                    </a>
                  </td>
                  <td>Vận hành hệ thống, duyệt danh mục, cấu hình thanh toán</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section id="tai-khoan-demo" className={styles.section}>
          <h2>Tài khoản demo</h2>
          <h3>Khách hàng</h3>
          <CredTable
            rows={[
              {
                role: 'Buyer Demo',
                email: 'buyer@ecomerce.local',
                password: 'Buyer@123456',
                note: 'Nên dùng khi thử mua hàng',
              },
              { role: 'Nguyễn An', email: 'an@ecomerce.local', password: 'Buyer@123456' },
              { role: 'Trần Bình', email: 'binh@ecomerce.local', password: 'Buyer@123456' },
              { role: 'Lê Chi', email: 'chi@ecomerce.local', password: 'Buyer@123456' },
            ]}
          />
          <h3>Cửa hàng</h3>
          <CredTable
            rows={[
              {
                role: 'Shop Demo',
                email: 'shop@ecomerce.local',
                password: 'Shop@123456',
                note: 'Cửa hàng chính để demo',
              },
              { role: 'Fashion House', email: 'fashion@ecomerce.local', password: 'Shop@123456' },
              { role: 'Tech Store', email: 'tech@ecomerce.local', password: 'Shop@123456' },
              { role: 'Home Living', email: 'home@ecomerce.local', password: 'Shop@123456' },
            ]}
          />
          <h3>Quản trị</h3>
          <CredTable
            rows={[
              {
                role: 'Admin Demo',
                email: 'admin@ecomerce.local',
                password: 'Admin@123456',
                note: 'Tài khoản quản trị chính',
              },
              { role: 'Ops Admin', email: 'ops@ecomerce.local', password: 'Admin@123456' },
            ]}
          />
        </section>

        <section id="web-khach-hang" className={styles.section}>
          <h2>Web khách hàng</h2>
          <p>
            Dùng để duyệt sản phẩm, đặt hàng và theo dõi đơn tại{' '}
            <a href={WEB} target="_blank" rel="noreferrer">
              {WEB}
            </a>
            .
          </p>
        </section>

        <section id="web-dang-nhap" className={styles.section}>
          <h3>Đăng nhập & tài khoản</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href="/login">Đăng nhập</FeatureLink> bằng tài khoản khách ở mục trên
              (form thường điền sẵn Buyer Demo).
            </li>
            <li>
              <FeatureLink href="/profile">Tài khoản</FeatureLink> — cập nhật tên hiển thị hoặc đổi
              mật khẩu.
            </li>
          </ul>
        </section>

        <section id="web-mua-hang" className={styles.section}>
          <h3>Mua hàng</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href="/">Sản phẩm</FeatureLink> — xem danh sách, lọc theo danh mục, mở
              chi tiết và thêm vào giỏ.
            </li>
            <li>Vào trang cửa hàng từ sản phẩm để xem các mặt hàng cùng shop.</li>
            <li>
              <FeatureLink href="/cart">Giỏ hàng</FeatureLink> — chỉnh số lượng rồi sang đặt hàng.
            </li>
            <li>
              <FeatureLink href="/checkout">Đặt hàng</FeatureLink> — chọn địa chỉ giao, phương thức
              thanh toán: khi nhận hàng (COD), thẻ nội địa, hoặc thẻ quốc tế.
            </li>
          </ul>
          <p className={styles.note}>
            Thanh toán thẻ chỉ hiện khi quản trị đã bật cổng thanh toán. COD luôn dùng được.
          </p>
        </section>

        <section id="web-dia-chi" className={styles.section}>
          <h3>Địa chỉ giao hàng</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href="/addresses">Sổ địa chỉ</FeatureLink> — thêm, sửa, xoá địa chỉ giao
              hàng (chọn quốc gia → tỉnh/TP → phường/xã).
            </li>
            <li>
              Tên người nhận và số điện thoại nhập lúc đặt hàng, không lưu trong sổ địa chỉ.
            </li>
          </ul>
        </section>

        <section id="web-don-hang" className={styles.section}>
          <h3>Đơn hàng & thanh toán</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href="/orders">Đơn hàng của tôi</FeatureLink> — xem lịch sử và mở chi
              tiết từng đơn.
            </li>
            <li>
              Trong chi tiết đơn: theo dõi trạng thái / vận chuyển; nếu thanh toán online chưa xong
              hoặc thất bại có nút <strong>Thanh toán lại</strong>.
            </li>
          </ul>
          <div className={styles.callout}>
            <p>
              <strong>Khi thanh toán online</strong>
            </p>
            <ol>
              <li>
                Đặt hàng → đơn ở trạng thái <em>Chờ thanh toán</em>
              </li>
              <li>
                Thanh toán thành công → đơn <em>Mới</em> (cửa hàng có thể xác nhận)
              </li>
              <li>
                Thanh toán thất bại / huỷ → đơn <em>Huỷ</em>; khách có thể thanh toán lại
              </li>
            </ol>
          </div>
        </section>

        <section id="merchant" className={styles.section}>
          <h2>Cổng cửa hàng (Merchant)</h2>
          <p>
            Dành cho chủ shop quản lý gian hàng tại{' '}
            <a href={MERCHANT} target="_blank" rel="noreferrer">
              {MERCHANT}
            </a>
            .
          </p>
        </section>

        <section id="merchant-dang-nhap" className={styles.section}>
          <h3>Đăng nhập & gian hàng</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${MERCHANT}/login`} external>
                Đăng nhập cửa hàng
              </FeatureLink>{' '}
              bằng tài khoản Shop Demo (hoặc shop khác ở mục Tài khoản demo).
            </li>
            <li>
              <FeatureLink href={`${MERCHANT}/`} external>
                Dashboard
              </FeatureLink>{' '}
              — xem nhanh số sản phẩm và đơn.
            </li>
            <li>
              <FeatureLink href={`${MERCHANT}/profile`} external>
                Gian hàng
              </FeatureLink>{' '}
              — cập nhật tên shop, ảnh đại diện, địa chỉ cửa hàng, mật khẩu.
            </li>
          </ul>
        </section>

        <section id="merchant-san-pham" className={styles.section}>
          <h3>Quản lý sản phẩm</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${MERCHANT}/products`} external>
                Danh sách sản phẩm
              </FeatureLink>{' '}
              — xem / mở chi tiết / sửa.
            </li>
            <li>
              <FeatureLink href={`${MERCHANT}/products/new`} external>
                Tạo sản phẩm
              </FeatureLink>{' '}
              — nhập giá, tồn kho, ảnh, gắn danh mục (có thể tạo danh mục mới và chờ admin duyệt).
            </li>
          </ul>
        </section>

        <section id="merchant-don-hang" className={styles.section}>
          <h3>Xử lý đơn hàng</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${MERCHANT}/orders`} external>
                Đơn hàng
              </FeatureLink>{' '}
              — lọc theo trạng thái và mở chi tiết.
            </li>
            <li>
              Với đơn <em>Mới</em>: xác nhận để bắt đầu giao, hoặc huỷ kèm lý do.
            </li>
            <li>
              Đơn đang <em>Chờ thanh toán</em> chưa xác nhận được — đợi khách thanh toán xong.
            </li>
          </ul>
        </section>

        <section id="admin" className={styles.section}>
          <h2>Trang quản trị (Admin)</h2>
          <p>
            Dành cho vận hành hệ thống tại{' '}
            <a href={ADMIN} target="_blank" rel="noreferrer">
              {ADMIN}
            </a>
            .
          </p>
        </section>

        <section id="admin-dang-nhap" className={styles.section}>
          <h3>Đăng nhập & tổng quan</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${ADMIN}/login`} external>
                Đăng nhập quản trị
              </FeatureLink>{' '}
              bằng Admin Demo.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/`} external>
                Overview
              </FeatureLink>{' '}
              — xem số liệu tổng quan.
            </li>
          </ul>
        </section>

        <section id="admin-quan-ly" className={styles.section}>
          <h3>Người dùng & danh mục</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${ADMIN}/users`} external>
                Users
              </FeatureLink>{' '}
              — tạo / sửa / xoá tài khoản khách.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/merchants`} external>
                Merchants
              </FeatureLink>{' '}
              — quản lý cửa hàng.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/categories`} external>
                Categories
              </FeatureLink>{' '}
              — duyệt hoặc từ chối danh mục do shop tạo.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/products`} external>
                Products
              </FeatureLink>{' '}
              — xem / chỉnh sản phẩm trên toàn hệ thống.
            </li>
          </ul>
        </section>

        <section id="admin-don-hang" className={styles.section}>
          <h3>Đơn hàng & giao hàng</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${ADMIN}/orders`} external>
                Orders
              </FeatureLink>{' '}
              — theo dõi đơn và lịch sử trạng thái.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/delivery-simulator`} external>
                Delivery sim
              </FeatureLink>{' '}
              — giả lập các bước giao hàng (tiếp nhận, đang giao, đã giao, hoàn hàng…).
            </li>
          </ul>
        </section>

        <section id="admin-thanh-toan" className={styles.section}>
          <h3>Thanh toán online</h3>
          <ul className={styles.featureList}>
            <li>
              <FeatureLink href={`${ADMIN}/payments`} external>
                Payments
              </FeatureLink>{' '}
              — bật thanh toán thẻ nội địa / quốc tế; có nút <strong>Điền demo</strong> để dùng cấu
              hình sandbox có sẵn.
            </li>
            <li>
              <FeatureLink href={`${ADMIN}/payment-callbacks`} external>
                Payment IPN
              </FeatureLink>{' '}
              — xem lại các phản hồi thanh toán từ cổng (thành công / thất bại) sau khi khách thanh
              toán.
            </li>
          </ul>
        </section>

        <section id="onepay" className={styles.section}>
          <h2>Thẻ test thanh toán</h2>
          <p>
            Khi checkout chọn thẻ nội địa hoặc thẻ quốc tế, dùng thông tin thẻ sandbox bên dưới trên
            trang thanh toán của OnePay.
          </p>
          <h3>Thẻ ATM nội địa</h3>
          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Ngân hàng</th>
                  <th>Số thẻ</th>
                  <th>Tên chủ thẻ</th>
                  <th>Hiệu lực</th>
                  <th>OTP</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>ABB</td>
                  <td>
                    <code>9704250000000001</code>
                  </td>
                  <td>NGUYEN VAN A</td>
                  <td>01/13</td>
                  <td>
                    <code>123456</code>
                  </td>
                </tr>
                <tr>
                  <td>VCB</td>
                  <td>
                    <code>9704360000000000002</code>
                  </td>
                  <td>NGUYEN VAN A</td>
                  <td>01/13</td>
                  <td>
                    <code>123456</code>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <h3>Thẻ quốc tế</h3>
          <div className={styles.tableWrap}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Loại</th>
                  <th>Số thẻ</th>
                  <th>Tên chủ thẻ</th>
                  <th>Hết hạn</th>
                  <th>CVV</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>Visa</td>
                  <td>
                    <code>4000000000000002</code>
                  </td>
                  <td>NGUYEN VAN A</td>
                  <td>05/24</td>
                  <td>
                    <code>123</code>
                  </td>
                </tr>
                <tr>
                  <td>Master</td>
                  <td>
                    <code>5313581000123430</code>
                  </td>
                  <td>NGUYEN VAN A</td>
                  <td>05/24</td>
                  <td>
                    <code>123</code>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <p className={styles.note}>
            Nếu trang ngân hàng hỏi mật khẩu: dùng <code>1234</code>.
          </p>
        </section>

        <section id="luong-demo" className={styles.section}>
          <h2>Luồng demo nhanh</h2>
          <ol className={styles.steps}>
            <li>
              Admin →{' '}
              <FeatureLink href={`${ADMIN}/payments`} external>
                Payments
              </FeatureLink>{' '}
              → Điền demo cho thẻ nội địa / quốc tế → Lưu.
            </li>
            <li>
              Web →{' '}
              <FeatureLink href="/login">đăng nhập khách</FeatureLink> → thêm sản phẩm → đặt hàng →
              chọn COD hoặc thanh toán thẻ.
            </li>
            <li>
              Nếu thanh toán thẻ: nhập thẻ test ở mục trên → hoàn tất → kiểm tra{' '}
              <FeatureLink href="/orders">đơn hàng</FeatureLink>.
            </li>
            <li>
              Merchant →{' '}
              <FeatureLink href={`${MERCHANT}/orders`} external>
                Đơn hàng
              </FeatureLink>{' '}
              → xác nhận đơn <em>Mới</em>.
            </li>
            <li>
              Admin (tuỳ chọn) →{' '}
              <FeatureLink href={`${ADMIN}/delivery-simulator`} external>
                Delivery sim
              </FeatureLink>{' '}
              để đẩy tiến trình giao hàng; hoặc{' '}
              <FeatureLink href={`${ADMIN}/payment-callbacks`} external>
                Payment IPN
              </FeatureLink>{' '}
              để xem lại kết quả thanh toán.
            </li>
          </ol>
        </section>
      </article>
    </div>
  )
}
